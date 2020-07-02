// +build !darwin,!netbsd,!freebsd,!openbsd,!dragonfly,!linux,!windows

package net

import nt "net"

type server struct {
	subEventLoopSet loadBalancer // event-loops for handling events
}

type eventloop struct {
	connCount int32 // number of active connections in event-loop
}

type listener struct {
	ln            nt.Listener
	pconn         nt.PacketConn
	lnaddr        nt.Addr
	addr, network string
}

func (ln *listener) renormalize() error {
	return nil
}

func (ln *listener) close() {
}

func serve(eventHandler EventHandler, listeners *listener, options *Options) error {
	return ErrUnsupportedPlatform
}
