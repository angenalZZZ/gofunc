// +build linux

package net

import "github.com/angenalZZZ/gofunc/net/internal/netpoll"

func (el *eventloop) handleEvent(fd int, ev uint32) error {
	if c, ok := el.connections[fd]; ok {
		switch c.outboundBuffer.IsEmpty() {
		// Don't change the ordering of processing EPOLLOUT | EPOLLRDHUP / EPOLLIN unless you're 100%
		// sure what you're doing!
		// Re-ordering can easily introduce bugs and bad side-effects, as I found out painfully in the past.
		case false:
			if ev&netpoll.OutEvents != 0 {
				return el.loopWrite(c)
			}
			return nil
		case true:
			if ev&netpoll.InEvents != 0 {
				return el.loopRead(c)
			}
			return nil
		}
	}
	return el.loopAccept(fd)
}
