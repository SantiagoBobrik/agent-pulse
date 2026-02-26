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
	http *http.Server
	hub  *Hub
	port int
}

func NewServer(hub *Hub, port int) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.Recoverer)

	r.Post("/event", handleEvent(hub))
	r.Get("/ws", serveWs(hub))

	return &Server{
		http: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           r,
			ReadHeaderTimeout: 5 * time.Second,
		},
		hub:  hub,
		port: port,
	}
}

func (s *Server) Start() error {
	err := s.http.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		if strings.Contains(err.Error(), "address already in use") {
			return fmt.Errorf("port %d is already in use. Change the port in ~/.config/claude-pulse/config.yaml", s.port)
		}
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.hub.Shutdown()
	return s.http.Shutdown(ctx)
}
