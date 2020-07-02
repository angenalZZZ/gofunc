// +build !linux,!darwin,!dragonfly,!freebsd,!netbsd,!windows

package netpoll

import (
	"errors"
	nt "net"
)

// SetKeepAlive sets the keepalive for the connection.
func SetKeepAlive(fd, secs int) error {
	// OpenBSD has no user-settable per-socket TCP keepalive options.
	return nil
}

// ReusePortListenPacket returns a net.PacketConn for UDP.
func ReusePortListenPacket(proto, addr string) (nt.PacketConn, error) {
	return nil, errors.New("SO_REUSEPORT/SO_REUSEADDR is not supported on this platform")
}

// ReusePortListen returns a net.Listener for TCP.
func ReusePortListen(proto, addr string) (nt.Listener, error) {
	return nil, errors.New("SO_REUSEPORT/SO_REUSEADDR is not supported on this platform")
}
