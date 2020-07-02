// +build linux darwin netbsd freebsd openbsd dragonfly

package net

import "golang.org/x/sys/unix"

func (svr *server) acceptNewConnection(fd int) error {
	nfd, sa, err := unix.Accept(fd)
	if err != nil {
		if err == unix.EAGAIN {
			return nil
		}
		return err
	}
	if err := unix.SetNonblock(nfd, true); err != nil {
		return err
	}
	el := svr.subEventLoopSet.next(nfd)
	c := newTCPConn(nfd, el, sa)
	_ = el.poller.Trigger(func() (err error) {
		if err = el.poller.AddRead(nfd); err != nil {
			return
		}
		el.connections[nfd] = c
		el.calibrateCallback(el, 1)
		err = el.loopOpen(c)
		return
	})
	return nil
}
