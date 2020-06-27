// 1. In a terminal trigger a graceful server restart (using the pid from your output):
//  > kill -USR2 pid
// 2. Run with supervisor
//  > command = /xxx/pid.py /xxx/log.pid /xxx/server
//  > kill -USR2 {pid.py pid}

package rpc

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/configfile"
	"github.com/angenalZZZ/gofunc/log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/facebookgo/grace/gracenet"
	"google.golang.org/grpc"
)

var (
	SIGUSR2 = syscall.Signal(0x1f) // ReStart Process
)

// GraceGrpc is used to wrap a grpc server that can be gracefully terminated & restarted
type GraceGrpc struct {
	server   *grpc.Server
	grace    *gracenet.Net
	listener net.Listener
	errors   chan error
	pidPath  string
	*log.Logger
}

// NewGraceGrpc is used to construct a new GraceGrpc
func NewGraceGrpc(s *grpc.Server, net, addr, pidPath string, logCfgFile string) (*GraceGrpc, error) {
	if logCfgFile == "" {
		logCfgFile = "log.yaml"
	}
	logCfg := new(log.AConfig)
	if err := configfile.YamlTo(logCfgFile, logCfg); err != nil {
		_ = fmt.Errorf("%s\n", err.Error())
	}

	gr := &GraceGrpc{
		server:  s,
		grace:   &gracenet.Net{},
		errors:  make(chan error),
		pidPath: pidPath,
		Logger:  log.Init(logCfg.Log),
	}
	listener, err := gr.grace.Listen(net, addr)
	if err != nil {
		return nil, err
	}
	gr.listener = listener
	return gr, nil
}

func (gr *GraceGrpc) startServe() {
	if err := gr.server.Serve(gr.listener); err != nil {
		gr.errors <- err
	}
}

func (gr *GraceGrpc) handleSignal() <-chan struct{} {
	terminate := make(chan struct{})
	go func() {
		ch := make(chan os.Signal, 10)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, SIGUSR2)
		for {
			sig := <-ch
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				signal.Stop(ch)
				gr.server.GracefulStop()
				close(terminate)
				return
			case SIGUSR2:
				if _, err := gr.grace.StartProcess(); err != nil {
					gr.errors <- err
				}
			}
		}
	}()
	return terminate
}

// storePid is used to write out PID to pidPath
func (gr *GraceGrpc) storePid(pid int) error {
	pidPath := gr.pidPath
	if pidPath == "" {
		return fmt.Errorf("no pid file path: %s", pidPath)
	}

	pidFile, err := os.OpenFile(pidPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("could not open pid file: %v", err)
	}
	defer pidFile.Close()

	_, err = pidFile.WriteString(fmt.Sprintf("%d", pid))
	if err != nil {
		return fmt.Errorf("could not write to pid file: %s", err)
	}
	return nil
}

// Serve is used to start grpc server.
// Serve will gracefully terminated or restarted when handling signals.
func (gr *GraceGrpc) Serve() error {
	if gr.listener == nil {
		return fmt.Errorf("gracegrpc must construct by new\n")
	}

	pid := os.Getpid()
	addrString := gr.listener.Addr().String()

	gr.Info().Msgf("Serving %s with pid %d\n", addrString, pid)

	if err := gr.storePid(pid); err != nil {
		return err
	}

	go gr.startServe()

	terminate := gr.handleSignal()

	select {
	case err := <-gr.errors:
		return err
	case <-terminate:
		gr.Error().Msgf("Exiting pid %d", os.Getpid())
		return nil
	}
}
