// +build darwin netbsd freebsd openbsd dragonfly

package net

import "github.com/angenalZZZ/gofunc/net/internal/netpoll"

func (el *eventloop) handleEvent(fd int, filter int16) error {
	if c, ok := el.connections[fd]; ok {
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
	return el.loopAccept(fd)
}
