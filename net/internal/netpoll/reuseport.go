// +build linux darwin netbsd freebsd openbsd dragonfly windows

package netpoll

import (
	nt "net"

	"github.com/libp2p/go-reuseport"
)

// ReusePortListenPacket returns a net.PacketConn for UDP.
func ReusePortListenPacket(proto, addr string) (nt.PacketConn, error) {
	return reuseport.ListenPacket(proto, addr)
}

// ReusePortListen returns a net.Listener for TCP.
func ReusePortListen(proto, addr string) (nt.Listener, error) {
	return reuseport.Listen(proto, addr)
}
