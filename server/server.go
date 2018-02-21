package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/asdine/storm"
	"github.com/go-chi/chi"
	"github.com/kochman/hotshots/config"
	"github.com/kochman/hotshots/log"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"golang.org/x/sys/unix"
)

/*
 * Server struct
 */
type Server struct {
	cfg     config.Config
	db      *storm.DB
	handler http.Handler
}

func New(cfg *config.Config) (*Server, error) {
	s := &Server{
		cfg: *cfg,
	}

	if err := s.Setup(); err != nil {
		return nil, err
	}

	router := chi.NewRouter()
	router.NotFound(NotFound)
	router.Route("/photos", func(router chi.Router) {
		router.Get("/", s.GetPhotos)
		router.Post("/", s.PostPhoto)
		router.Get("/ids", s.GetPhotoIDs)
		router.Route("/{pid}", func(router chi.Router) {
			router.Use(s.PhotoCtx)
			router.Get("/image.jpg", s.GetPhoto)
			router.Get("/thumb.jpg", s.GetThumbnail)
			router.Get("/meta", s.GetPhotoMetadata)
		})
	})

	s.handler = router
	return s, nil
}

func (s *Server) Setup() error {
	if !CanAccessDirectory(s) {
		return errors.New(fmt.Sprintf("Directory %s not accessible", s.cfg.PhotosDirectory))
	}

	if _, err := os.Stat(s.cfg.ImgFolder()); err != nil {
		os.Mkdir(s.cfg.ImgFolder(), 0775)
	}

	if _, err := os.Stat(s.cfg.ConfFolder()); err != nil {
		os.Mkdir(s.cfg.ConfFolder(), 0775)
	}

	db, err := storm.Open(s.cfg.StormFile())
	if err != nil {
		return err
	}
	s.db = db

	if err := s.db.Init(&Photo{}); err != nil {
		return err
	}

	exif.RegisterParsers(mknote.All...)

	return nil
}

func CanAccessDirectory(serv *Server) bool {
	stat, statErr := os.Stat(serv.cfg.PhotosDirectory)
	if statErr != nil || !stat.IsDir() {
		return false
	}
	readErr := unix.Access(serv.cfg.PhotosDirectory, unix.R_OK)
	writeErr := unix.Access(serv.cfg.PhotosDirectory, unix.W_OK)
	return readErr == nil && writeErr == nil
}

func (s *Server) Run() {
	if err := http.ListenAndServe(s.cfg.ListenURL, s.handler); err != nil {
		log.WithError(err).Error("Unable to serve")
	}
}
