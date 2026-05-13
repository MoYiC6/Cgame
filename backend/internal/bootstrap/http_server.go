package bootstrap

import (
	"context"
	stderrors "errors"
	"net"
	"net/http"
)

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(addr string, handler http.Handler) *HTTPServer {
	return &HTTPServer{server: &http.Server{Addr: addr, Handler: handler}}
}

func (s *HTTPServer) Run() error {
	if s == nil || s.server == nil {
		return nil
	}
	err := s.server.ListenAndServe()
	if stderrors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *HTTPServer) Serve(listener net.Listener) error {
	if s == nil || s.server == nil {
		return nil
	}
	err := s.server.Serve(listener)
	if stderrors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	if s == nil || s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}
