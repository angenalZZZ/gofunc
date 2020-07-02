// +build windows

package net

import (
	"github.com/klauspost/crc32"
	"time"

	"github.com/angenalZZZ/gofunc/net/pool/bytebuffer"
)

// hashCode hashes a string to a unique hashcode.
func hashCode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	return -v
}

func (svr *server) listenerRun() {
	var err error
	defer func() { svr.signalShutdown(err) }()
	var packet [0x10000]byte
	for {
		if svr.ln.pconn != nil {
			// Read data from UDP socket.
			n, addr, e := svr.ln.pconn.ReadFrom(packet[:])
			if e != nil {
				err = e
				return
			}
			buf := bytebuffer.Get()
			_, _ = buf.Write(packet[:n])

			el := svr.subEventLoopSet.next(hashCode(addr.String()))
			el.ch <- &udpIn{newUDPConn(el, svr.ln.lnaddr, addr, buf)}
		} else {
			// Accept TCP socket.
			conn, e := svr.ln.ln.Accept()
			if e != nil {
				err = e
				return
			}
			el := svr.subEventLoopSet.next(hashCode(conn.RemoteAddr().String()))
			c := newTCPConn(conn, el)
			el.ch <- c
			go func() {
				var packet [0x10000]byte
				for {
					n, err := c.conn.Read(packet[:])
					if err != nil {
						_ = c.conn.SetReadDeadline(time.Time{})
						el.ch <- &stderr{c, err}
						return
					}
					buf := bytebuffer.Get()
					_, _ = buf.Write(packet[:n])
					el.ch <- &tcpIn{c, buf}
				}
			}()
		}
	}
}
