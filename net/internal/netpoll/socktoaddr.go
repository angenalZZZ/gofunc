// +build linux darwin netbsd freebsd openbsd dragonfly

package netpoll

import (
	nt "net"

	"golang.org/x/sys/unix"
)

// SockaddrToTCPOrUnixAddr converts a Sockaddr to a net.TCPAddr or net.UnixAddr.
// Returns nil if conversion fails.
func SockaddrToTCPOrUnixAddr(sa interface{}) nt.Addr {
	switch sa := sa.(type) {
	case *unix.SockaddrInet4:
		ip := sockaddrInet4ToIP(sa)
		return &nt.TCPAddr{IP: ip, Port: sa.Port}
	case *unix.SockaddrInet6:
		ip, zone := sockaddrInet6ToIPAndZone(sa)
		return &nt.TCPAddr{IP: ip, Port: sa.Port, Zone: zone}
	case *unix.SockaddrUnix:
		return &nt.UnixAddr{Name: sa.Name, Net: "unix"}
	}
	return nil
}

// SockaddrToUDPAddr converts a Sockaddr to a net.UDPAddr
// Returns nil if conversion fails.
func SockaddrToUDPAddr(sa interface{}) *nt.UDPAddr {
	switch sa := sa.(type) {
	case *unix.SockaddrInet4:
		ip := sockaddrInet4ToIP(sa)
		return &nt.UDPAddr{IP: ip, Port: sa.Port}
	case *unix.SockaddrInet6:
		ip, zone := sockaddrInet6ToIPAndZone(sa)
		return &nt.UDPAddr{IP: ip, Port: sa.Port, Zone: zone}
	}
	return nil
}

// sockaddrInet4ToIPAndZone converts a SockaddrInet4 to a net.IP.
// It returns nil if conversion fails.
func sockaddrInet4ToIP(sa *unix.SockaddrInet4) nt.IP {
	ip := make([]byte, 16)
	// V4InV6Prefix
	ip[10] = 0xff
	ip[11] = 0xff
	copy(ip[12:16], sa.Addr[:])
	return ip
}

// sockaddrInet6ToIPAndZone converts a SockaddrInet6 to a net.IP with IPv6 Zone.
// It returns nil if conversion fails.
func sockaddrInet6ToIPAndZone(sa *unix.SockaddrInet6) (nt.IP, string) {
	ip := make([]byte, 16)
	copy(ip, sa.Addr[:])
	return ip, ip6ZoneToString(int(sa.ZoneId))
}

// ip6ZoneToString converts an IP6 Zone unix int to a net string
// returns "" if zone is 0
func ip6ZoneToString(zone int) string {
	if zone == 0 {
		return ""
	}
	if ifi, err := nt.InterfaceByIndex(zone); err == nil {
		return ifi.Name
	}
	return int2decimal(uint(zone))
}

// Convert int to decimal string.
func int2decimal(i uint) string {
	if i == 0 {
		return "0"
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; i > 0; i /= 10 {
		bp--
		b[bp] = byte(i%10) + '0'
	}
	return string(b[bp:])
}
