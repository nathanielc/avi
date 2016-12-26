package server

import (
	"net"
	"net/http"
)

type ServerConf struct {
	Addr    string
	DataDir string
}

type Server struct {
	c ServerConf
	l *net.TCPListener
	h *Handler
}

func New(c ServerConf) (*Server, error) {
	d, err := newData(c.DataDir)
	if err != nil {
		return nil, err
	}
	h := newHandler(d)
	return &Server{
		c: c,
		h: h,
	}, nil
}

func (s *Server) Serve() error {
	addr, err := net.ResolveTCPAddr("tcp4", s.c.Addr)
	if err != nil {
		return err
	}
	l, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		return err
	}
	s.l = l
	srv := &http.Server{
		Handler: s.h,
	}
	if err := s.h.Open(); err != nil {
		return err
	}
	return srv.Serve(s.l)
}

func (s *Server) Close() {
	s.l.Close()
}
