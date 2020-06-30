//go:generate protoc -I ../../../../../rpc/proto --go_out=plugins=grpc:. ../../../../../rpc/proto/cache.proto
//!go:generate protoc -I ../../../../../rpc/proto --csharp_out=. ../../../../../rpc/proto/cache.proto
// go get github.com/angenalZZZ/gofunc/data/cache/fastcache/cmd/fastcache1e6
// go build -ldflags "-s -w" -o A:/test/ .
// cd A:/test/ && fastcache1e6 -c 2 -d 128 -t 10000000
// cache1.benchmark(4CPU+16G+MHD) GET:2000M/s SET:200M/s FLUSH:0.4s
// cache2.benchmark(8CPU+16G+SSD) GET:2000M/s SET:200M/s FLUSH:0.1s
// buntdb-benchmark(8CPU+16G+SSD) GET:500M/s  SET:23M/s

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
