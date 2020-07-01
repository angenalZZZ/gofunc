package main

import (
	"fmt"
	"github.com/panjf2000/gnet"
)

type netTcpServer struct {
	*gnet.EventServer
}

func (es *netTcpServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	fmt.Printf("server is listening on %s (multi-cores: %t, loops: %d)\n",
		srv.Addr.String(), srv.Multicore, srv.NumEventLoop)
	return
}

func (es *netTcpServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	// Write synchronously.
	out = frame
	return

	/*
		// Write asynchronously.
		data := append([]byte{}, frame...)
		go func() {
			time.Sleep(time.Second)
			c.AsyncWrite(data)
		}()
		return
	*/
}

func ProdTCP() {
	echo := new(netTcpServer)
	if err := gnet.Serve(echo, fmt.Sprintf("tcp://:%d", *flagPort), gnet.WithMulticore(true)); err != nil {
		_ = fmt.Errorf("failed to serve: %v\n", err)
	}
}
