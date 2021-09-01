package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

type Constructor func(ctx context.Context, group *errgroup.Group) Server

var OsExitError = errors.New("os exit signal")

// Server is an interface for implementing concurrent servers.
// CreateAndRunServer() implements a graceful shutdown waiting for all the resources to close.
//  It calls Init() and Start() on the server.
//
// Use BasicServer.Group to start goroutines
// When implementing an http server, use BasicHttpServer struct.
//
type Server interface {
	Init()
	Start()
	Shutdown(ctx context.Context) error
}

type BasicServer struct {
	Ctx   context.Context
	Group *errgroup.Group
}

func (b *BasicServer) Init() {
	// Empty implementation. Not all servers need this one.
}

// CreateAndRunServer creates, initializes and starts a Server.
// Also makes sure that all the subroutines are finished before exiting.
func CreateAndRunServer(constructor Constructor, gracefulTimeout time.Duration) (err error) {
	// Server cancellation
	serverContext, serverCancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(serverContext)
	// Used to track goroutines
	s := constructor(ctx, group)
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
		select {
		case <-serverContext.Done():
			fmt.Println("Context closed")
		case sig := <-osSig:
			fmt.Printf("Received OS signal: %v\n", sig)
			err = OsExitError
		}
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
	if err != nil {
		_ = group.Wait()
		return err
	}
	return group.Wait()
}
