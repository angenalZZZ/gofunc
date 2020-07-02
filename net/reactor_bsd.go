// +build darwin netbsd freebsd openbsd dragonfly

package net

import "github.com/angenalZZZ/gofunc/net/internal/netpoll"

func (svr *server) activateMainReactor() {
	defer svr.signalShutdown()

	svr.logger.Printf("main reactor exits with error:%v\n", svr.mainLoop.poller.Polling(func(fd int, filter int16) error {
		return svr.acceptNewConnection(fd)
	}))
}

func (svr *server) activateSubReactor(el *eventloop) {
	defer func() {
		el.closeAllConns()
		if el.idx == 0 && svr.opts.Ticker {
			close(svr.ticktock)
		}
		svr.signalShutdown()
	}()

	if el.idx == 0 && svr.opts.Ticker {
		go el.loopTicker()
	}

	svr.logger.Printf("event-loop:%d exits with error:%v\n", el.idx, el.poller.Polling(func(fd int, filter int16) error {
		if c, ack := el.connections[fd]; ack {
			if filter == netpoll.EVFilterSock {
				return el.loopCloseConn(c, nil)
			}
			switch c.outboundBuffer.IsEmpty() {
			// Don't change the ordering of processing EVFILT_WRITE | EVFILT_READ | EV_ERROR/EV_EOF unless you're 100%
			// sure what you're doing!
			// Re-ordering can easily introduce bugs and bad side-effects, as I found out painfully in the past.
			case false:
				if filter == netpoll.EVFilterWrite {
					return el.loopWrite(c)
				}
				return nil
			case true:
				if filter == netpoll.EVFilterRead {
					return el.loopRead(c)
				}
				return nil
			}
		}
		return nil
	}))
}
