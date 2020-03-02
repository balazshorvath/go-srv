package server

import (
	"context"
	"net/http"
)

// BasicHttpServer implements the Start() and Shutdown() functions of the Server interface.
// Implement Init() to initialize the routing and the constructor to instantiate dependencies.
//
type BasicHttpServer struct {
	BasicServer
	Srv *http.Server
}

func (b *BasicHttpServer) Start() {
	// Server routine
	go func() {
		b.Group.Add(1)
		defer b.Group.Done()
		// Start
		_ = b.Srv.ListenAndServe()
	}()
}

func (b *BasicHttpServer) Shutdown(ctx context.Context) error {
	return b.Srv.Shutdown(ctx)
}
