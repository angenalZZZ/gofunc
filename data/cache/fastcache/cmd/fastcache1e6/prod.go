package main

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/testdata"
)

func Prod() {
	var opts []grpc.ServerOption
	if *flagTls {
		if *flagCert == "" {
			*flagCert = testdata.Path("server1.pem")
		}
		if *flagKey == "" {
			*flagKey = testdata.Path("server1.key")
		}
		cred, err := credentials.NewServerTLSFromFile(*flagCert, *flagKey)
		if err != nil {
			_ = fmt.Errorf("failed to generate credentials %v\n", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(cred)}
	}

	svr := grpc.NewServer(opts...)
	RegisterCacheServiceServer(svr, &cacheServiceImpl{})
	reflection.Register(svr)

	g, err := rpc.NewGraceGrpc(svr, "tcp", fmt.Sprintf("%d", *flagPort), "log.pid", "log.yaml")
	if err != nil {
		_ = fmt.Errorf("failed to new grace grpc: %v\n", err)
	}
	if err = g.Serve(); err != nil {
		_ = fmt.Errorf("failed to serve: %v\n", err)
	}
}
