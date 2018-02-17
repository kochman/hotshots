package server

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/asdine/storm"
	"github.com/go-chi/chi"
	"github.com/kochman/hotshots/config"
	"github.com/kochman/hotshots/log"
	"golang.org/x/sys/unix"
)

/*
 * Constants
 */

/*
 * Server struct
 */
type Server struct {
	cfg     config.Config
	db      *storm.DB
	handler http.Handler
}

/*
 * Data structs
 */

type Photo struct {
	ID 			string `storm:"id"`
	UploadedAt	time.Time `storm:"index"`
}

/*
 * Response Structs
 */

type GetPhotosResponse struct {
	Success bool
	Photos	[]Photo
}

type PostPhotoResponse struct {
	Success bool
	NewID 	string	`json:",omitempty"`
}

type GetPhotoMetadataResponse struct {
	Success bool
	Photo	Photo
}

type ErrorResponse struct {
	Success bool
	Error 	string
}

type NotFoundResponse struct {
	Success	bool
	Error	string
	URI		string
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
		router.Route("/{pid}", func(router chi.Router) {
			router.Use(s.PhotoCtx)
			router.Get("/image.jpg", s.GetPhoto)
			// router.Get("/thumb.jpg")
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

	return nil
}

func NotFound(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(404)
	v := NotFoundResponse{
		Success: false,
		Error: "Requested URI not found",
		URI: r.RequestURI,
	}
	WriteJsonResponse(v, w)
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

func (s *Server) GetPhotos(w http.ResponseWriter, r *http.Request) {
	var photos []Photo
	if err := s.db.All(&photos); err != nil{
		log.Error(err)
		WriteError("Unable to query photos", 500, w)
		return
	}

	v := GetPhotosResponse{
		Success: true,
		Photos: photos,
	}
	WriteJsonResponse(v, w)
}

func (s *Server) PostPhoto(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(1 << 23) // 8M max memory

	input, _, err := r.FormFile("photo")
	if err != nil {
		log.Info(err)
		WriteError("Unable to parse form value 'photo'", 400, w)
		return
	}

	id, err := GenPhotoID(input)
	if err != nil {
		log.Error(err)
		WriteError("Unable to generate photo ID", 500, w)
		return
	}

	var existingPhoto Photo
	if err := s.db.One("ID", id, &existingPhoto); err == nil {
		log.Info(err)
		WriteError("Photo already exists", 400, w)
		return
	} else if err != storm.ErrNotFound {
		log.Error(err)
		WriteError("Unable to query database", 500, w)
		return
	}

	photoPath := path.Join(s.cfg.ImgFolder(), fmt.Sprintf("%s.jpg", id))
	log.Info(fmt.Sprint("Saving image ", photoPath))
	output, err := os.OpenFile(photoPath, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil{
		log.Error(err)
		WriteError("Unable to write to output file", 500, w)
		return
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		log.Error(err)
		WriteError("Unable to copy input to output file", 500, w)
		return
	}

	newPhoto := Photo{
		ID: id,
		UploadedAt: time.Now(),
	}

	if err := s.db.Save(&newPhoto); err != nil {
		log.Error(err)
		WriteError("Unable to save to database", 500, w)
		os.Remove(photoPath)
		return
	}

	v := PostPhotoResponse{
		Success: true,
		NewID: id,
	}
	WriteJsonResponse(v, w)
}

func (s *Server) GetPhotoMetadata(w http.ResponseWriter, r *http.Request) {
	v := GetPhotoMetadataResponse{
		Photo: r.Context().Value("photo").(Photo),
		Success: true,
	}
	WriteJsonResponse(v, w)
}

func (s *Server) GetPhoto(w http.ResponseWriter, r *http.Request) {
	photo := r.Context().Value("photo").(Photo)
	photoPath := path.Join(s.cfg.ImgFolder(), fmt.Sprintf("%s.jpg", photo.ID))
	output, err := os.OpenFile(photoPath, os.O_RDONLY, 0660)
	if err != nil {
		log.Error(err)
		WriteError("Unable to access internal storage of image", 500, w)
		return
	}
	defer output.Close()

	w.Header().Set("Content-Type", "image/jpeg")
	if _, err := io.Copy(w, output); err != nil {
		log.Error(err)
		WriteError("Unable to return image data", 500, w)
		return
	}
}

func (s *Server) PhotoCtx (next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		photoID := chi.URLParam(r, "pid")
		var photo Photo
		if err := s.db.One("ID", photoID, &photo); err != nil {
			log.Info(err)
			WriteError("Unable to find photo id", 404, w)
			return
		}
		ctx := context.WithValue(r.Context(), "photo", photo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GenPhotoID(f io.ReadSeeker) (string, error) {
	digest := sha1.New()
	if _, err := io.Copy(digest, f); err != nil {
		return "", err
	}
	f.Seek(0, io.SeekStart)
	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}

func WriteError(s string, status int, w http.ResponseWriter){
	w.WriteHeader(status)
	v := ErrorResponse{
		Success: false,
		Error: s,
	}
	WriteJsonResponse(v, w)
}

func WriteJsonResponse(v interface{}, w http.ResponseWriter) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (s *Server) Run() {
	if err := http.ListenAndServe(s.cfg.ListenURL, s.handler); err != nil {
		log.WithError(err).Error("Unable to serve")
	}
}
