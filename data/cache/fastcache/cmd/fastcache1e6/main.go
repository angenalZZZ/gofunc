//go:generate protoc -I ../../../../../rpc/proto --go_out=plugins=grpc:. ../../../../../rpc/proto/cache.proto
//!go:generate protoc -I ../../../../../rpc/proto --csharp_out=. ../../../../../rpc/proto/cache.proto
// go get github.com/angenalZZZ/gofunc/data/cache/fastcache/cmd/fastcache1e6
// go build -ldflags "-s -w" -o A:/test/ .
// cd A:/test/ && fastcache1e6 -c 2 -d 128 -t 10000000
// 1.benchmark(4CPU+16G+MHD) GET:2000w/Qps SET:200w/Qps FLUSH:0.4s
// 2.benchmark(8CPU+16G+SSD) GET:2000w/Qps SET:200w/Qps FLUSH:0.1s
// >buntdb-benchmark(8CPU+16G+MHD) GET:500w/Qps SET:23w/Qps

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
