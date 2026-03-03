package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	http       *http.Server
	dispatcher *Dispatcher
	port       int
}

func NewServer(dispatcher *Dispatcher, broker *Broker, port int, bindAddress string) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.Recoverer)

	r.Post("/event", handleEvent(dispatcher))
	r.Get("/events/stream", handleSSE(broker))

	return &Server{
		http: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", bindAddress, port),
			Handler:           r,
			ReadHeaderTimeout: 5 * time.Second,
		},
		dispatcher: dispatcher,
		port:       port,
	}
}

func (s *Server) Start() error {
	err := s.http.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		if strings.Contains(err.Error(), "address already in use") {
			return fmt.Errorf("port %d is already in use. Change the port in ~/.config/agent-pulse/config.yaml", s.port)
		}
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
