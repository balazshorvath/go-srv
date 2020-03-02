package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Constructor func(ctx context.Context, group *sync.WaitGroup) Server

// Server is an interface for implementing concurrent servers.
// CreateAndRunServer() implements a graceful shutdown waiting for all the resources to close.
//  It calls Init() and Start() on the server.
//
// When implementing an http server, use BasicHttpServer struct.
//
type Server interface {
	Init()
	Start()
	Shutdown(ctx context.Context) error
}

type BasicServer struct {
	Ctx   context.Context
	Group *sync.WaitGroup
}

// CreateAndRunServer creates, initializes and starts a Server.
// Also makes sure that all the subroutines are finished before exiting.
func CreateAndRunServer(constructor Constructor, gracefulTimeout time.Duration) {
	// Used to track goroutines
	group := &sync.WaitGroup{}
	// Server cancellation
	serverContext, serverCancel := context.WithCancel(context.Background())
	s := constructor(serverContext, group)
	s.Init()
	s.Start()
	// Stop main thread until we want to shut down
	done := make(chan struct{})
	defer close(done)
	go func() {
		// Setup OS signals
		osSig := make(chan os.Signal)
		signal.Notify(osSig, syscall.SIGINT, syscall.SIGTERM)
		// Wait for input
		<-osSig
		// Cancel the server context (initiate shutdown)
		// Give the server some time to close resources
		serverCancel()
		ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			fmt.Printf("Server shutdown with error %v\n", err)
		}
		done <- struct{}{}
	}()
	<-done
	group.Wait()
}
