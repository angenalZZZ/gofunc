package main

import (
	"fmt"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pool/goroutine"
	"sync/atomic"
	"time"
)

type netTcpServer struct {
	*gnet.EventServer
	*goroutine.Pool
	connCount int32
}

// OnInitComplete fires when the server is ready for accepting connections.
// The server parameter has information and various utilities.
func (es *netTcpServer) OnInitComplete(server gnet.Server) (action gnet.Action) {
	fmt.Printf("server is listening on %s (multi-cores: %t, loops: %d)\n",
		server.Addr.String(), server.Multicore, server.NumEventLoop)
	return
}

// OnShutdown fires when the server is being shut down, it is called right after
// all event-loops and connections are closed.
func (es *netTcpServer) OnShutdown(server gnet.Server) {}

// OnOpened fires when a new connection has been opened.
// The info parameter has information about the connection such as
// it's local and remote address.
// Use the out return value to write data to the connection.
func (es *netTcpServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	_ = atomic.AddInt32(&es.connCount, 1)
	return
}

// OnClosed fires when a connection has been closed.
// The err parameter is the last known connection error.
func (es *netTcpServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	_ = atomic.AddInt32(&es.connCount, -1)
	return
}

// PreWrite fires just before any data is written to any client socket.
func (es *netTcpServer) PreWrite() {}

// React fires when a connection sends the server data.
// Invoke c.Read() or c.ReadN(n) within the parameter c to read incoming data from client/connection.
// Use the out return value to write data to the client/connection.
func (es *netTcpServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	// Write synchronously.
	out = defaultService.Handle(frame)
	return

	/*
		// Use ants pool to unblock the event-loop.
		_ = es.Pool.Submit(func() {
			c.AsyncWrite(frame)
		})
	*/

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

// Tick fires immediately after the server starts and will fire again
// following the duration specified by the delay return value.
func (es *netTcpServer) Tick() (delay time.Duration, action gnet.Action) {
	fmt.Println("tick, conn count:", es.connCount)
	delay = time.Second
	return
}

func ProdTCP() {
	echo := new(netTcpServer)
	echo.Pool = goroutine.Default()
	defer echo.Pool.Release()
	if err := gnet.Serve(echo, fmt.Sprintf("tcp://:%d", *flagPort),
		gnet.WithMulticore(true),
		gnet.WithTicker(true)); err != nil {
		_ = fmt.Errorf("failed to serve: %v\n", err)
	}
}
