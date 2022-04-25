package server

import (
	"context"
	"net/http"
	"time"
)

type Config struct {
	Addr              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	ReadHeaderTimeout time.Duration
	CancelTimeout     time.Duration
}

type Server struct {
	srv http.Server
	ct  time.Duration
}

func NewServer(c Config, h http.Handler) *Server {
	s := &Server{}
	s.ct = c.CancelTimeout
	s.srv = http.Server{
		Addr:              c.Addr,
		Handler:           h,
		ReadTimeout:       c.ReadTimeout,
		WriteTimeout:      c.WriteTimeout,
		ReadHeaderTimeout: c.ReadHeaderTimeout,
	}
	return s
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), s.ct)
	s.srv.Shutdown(ctx)
	cancel()
}

func (s *Server) Start() {
	go s.srv.ListenAndServe()
}
