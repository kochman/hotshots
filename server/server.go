package server

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/kochman/hotshots/config"
	"github.com/kochman/hotshots/log"
)

type Server struct {
	cfg     config.Config
	handler http.Handler
}

func New(cfg *config.Config) (*Server, error) {
	s := &Server{
		cfg: *cfg,
	}

	router := chi.NewRouter()
	s.handler = router
	return s, nil
}

func (s *Server) Run() {
	if err := http.ListenAndServe(s.cfg.ListenURL, s.handler); err != nil {
		log.WithError(err).Error("Unable to serve")
	}
}
