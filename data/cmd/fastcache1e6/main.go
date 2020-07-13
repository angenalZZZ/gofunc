//go:generate protoc -I ../../../../../rpc/proto --go_out=plugins=grpc:. ../../../../../rpc/proto/cache.proto
//!go:generate protoc -I ../../../../../rpc/proto --csharp_out=. ../../../../../rpc/proto/cache.proto
// go get github.com/angenalZZZ/gofunc/data/cache/fastcache/cmd/fastcache1e6
// go build -ldflags "-s -w" -o A:/test/ .

// TEST: fastcache1e6 -c 2 -d 128 -t 10000000
// cache1.benchmark(4CPU+16G+MHD) GET:20Mq/s SET:2Mq/s FLUSH:0.4s
// cache2.benchmark(8CPU+16G+SSD) GET:20Mq/s SET:2Mq/s FLUSH:0.1s
// buntdb-benchmark(8CPU+16G+SSD) GET:5Mq/s  SET:230Kq/s

// SHM: fastcache1e6 -prod=true -s=0 -d=1 -a=ipc://cache

// TCP: fastcache1e6 -prod=true -s=2 -p=6060
// CSharp.benchmark(4CPU+16G+MHD) GET:60Kq/s SET:60Kq/s

// WS: fastcache1e6 -prod=true -s=3 -p=6060

// NatS: fastcache1e6 -prod=true -s=4 -p=4222 -name=cache -token=HGJ766GR767FKJU0

package main

import (
	"flag"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		flag.Usage()
		return
	}

	if *flagProd == false {
		Stage()
	} else {
		Prod()
	}
}
