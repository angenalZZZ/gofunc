package f

import (
	"log"
	"net"
	"strings"
)

// IP get internal IPs.
func IP(prefixList ...string) (ip []string) {
	adders, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal("Oops: " + err.Error())
	}

	i := len(prefixList)
	for _, addr := range adders {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				s := ipNet.IP.String()
				if i == 0 {
					ip = append(ip, s)
					continue
				}
				for _, prefix := range prefixList {
					if strings.HasPrefix(s, prefix) {
						ip = append(ip, s)
					}
				}
			}
		}
	}
	return
}
