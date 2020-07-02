// +build linux netbsd freebsd dragonfly

package netpoll

import "golang.org/x/sys/unix"

// SetKeepAlive sets the keepalive for the connection.
func SetKeepAlive(fd, secs int) error {
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1); err != nil {
		return err
	}
	if err := unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, secs); err != nil {
		return err
	}
	return unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, secs)
}
