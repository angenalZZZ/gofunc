// +build linux darwin netbsd freebsd openbsd dragonfly

package net

import (
	nt "net"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

type listener struct {
	f             *os.File
	fd            int
	ln            nt.Listener
	once          sync.Once
	pconn         nt.PacketConn
	lnaddr        nt.Addr
	addr, network string
}

// renormalize takes the net listener and detaches it from it's parent
// event loop, grabs the file descriptor, and makes it non-blocking.
func (ln *listener) renormalize() error {
	var err error
	switch netln := ln.ln.(type) {
	case nil:
		switch pconn := ln.pconn.(type) {
		case *nt.UDPConn:
			ln.f, err = pconn.File()
		}
	case *nt.TCPListener:
		ln.f, err = netln.File()
	case *nt.UnixListener:
		ln.f, err = netln.File()
	}
	if err != nil {
		ln.close()
		return err
	}
	ln.fd = int(ln.f.Fd())
	return unix.SetNonblock(ln.fd, true)
}

func (ln *listener) close() {
	ln.once.Do(
		func() {
			if ln.f != nil {
				sniffErrorAndLog(ln.f.Close())
			}
			if ln.ln != nil {
				sniffErrorAndLog(ln.ln.Close())
			}
			if ln.pconn != nil {
				sniffErrorAndLog(ln.pconn.Close())
			}
			if ln.network == "unix" {
				sniffErrorAndLog(os.RemoveAll(ln.addr))
			}
		})
}
