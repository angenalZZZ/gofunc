// +build windows

package net

import (
	nt "net"
	"os"
	"sync"
)

type listener struct {
	ln            nt.Listener
	once          sync.Once
	pconn         nt.PacketConn
	lnaddr        nt.Addr
	addr, network string
}

func (ln *listener) renormalize() error {
	return nil
}

func (ln *listener) close() {
	ln.once.Do(func() {
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
