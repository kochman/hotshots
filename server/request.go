package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/asdine/storm"
	"github.com/go-chi/chi"
	"github.com/kochman/hotshots/log"
)

/*
 * Response Structs
 */

type GetPhotoIDsResponse struct {
	Success bool
	IDs     []string
}

type GetPhotosResponse struct {
	Success bool
	Photos  []Photo
}

type PostPhotoResponse struct {
	Success bool
	NewID   string `json:",omitempty"`
}

type GetPhotoMetadataResponse struct {
	Success bool
	Photo   Photo
}

type ErrorResponse struct {
	Success bool
	Error   string
}

type NotFoundResponse struct {
	Success bool
	Error   string
	URI     string
}

/*
 * Handlers
 */

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	v := NotFoundResponse{
		Success: false,
		Error:   "Requested URI not found",
		URI:     r.RequestURI,
	}
	WriteJsonResponse(v, w)
}

func (s *Server) PhotoCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) GetPhotos(w http.ResponseWriter, r *http.Request) {
	start, limit, err := GetPaginateValues(r)
	if err != nil {
		log.Error(err)
		WriteError("Unable to parse query string", 400, w)
		return
	}

	var photos []Photo
	if err := s.db.All(&photos, start, limit); err != nil {
		log.Error(err)
		WriteError("Unable to query photos", 500, w)
		return
	}

	v := GetPhotosResponse{
		Success: true,
		Photos:  photos,
	}
	WriteJsonResponse(v, w)
}

func (s *Server) GetPhotoIDs(w http.ResponseWriter, r *http.Request) {
	start, limit, err := GetPaginateValues(r)
	if err != nil {
		log.Error(err)
		WriteError("Unable to parse query string", 400, w)
		return
	}
	var photos []Photo
	if err := s.db.All(&photos, start, limit); err != nil {
		log.Error(err)
		WriteError("Unable to query photos", 500, w)
		return
	}
	ids := make([]string, len(photos))
	for i, photo := range photos {
		ids[i] = photo.ID
	}

	v := GetPhotoIDsResponse{
		Success: true,
		IDs:     ids,
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
	defer input.Close()

	id, err := GenPhotoID(input)
	if err != nil {
		log.Error(err)
		WriteError("Unable to generate photo ID", 500, w)
		return
	}

	var existingPhoto Photo
	if err := s.db.One("ID", id, &existingPhoto); err == nil {
		log.Info(fmt.Sprintf("Photo %s already exists", id))
		WriteError("Photo already exists", 400, w)
		return
	} else if err != storm.ErrNotFound {
		log.Error(err)
		WriteError("Unable to query database", 500, w)
		return
	}

	photoPath := path.Join(s.cfg.ImgFolder(), fmt.Sprintf("%s.jpg", id))
	thumbPath := path.Join(s.cfg.ImgFolder(), fmt.Sprintf("%s-thumb.jpg", id))

	xif, rect, err := ProcessPhoto(input, id, photoPath, thumbPath)
	if err != nil {
		log.Error("Failed to save image")
		WriteError("Unable to save image", 500, w)

		os.Remove(photoPath)
		os.Remove(thumbPath)
		return
	}

	newPhoto := NewPhoto(id, *rect, *xif)

	if err := s.db.Save(newPhoto); err != nil {
		log.Error(err)
		WriteError("Unable to save to database", 500, w)

		os.Remove(photoPath)
		os.Remove(thumbPath)
		return
	}

	v := PostPhotoResponse{
		Success: true,
		NewID:   id,
	}
	WriteJsonResponse(v, w)
}

func (s *Server) GetPhotoMetadata(w http.ResponseWriter, r *http.Request) {
	v := GetPhotoMetadataResponse{
		Photo:   r.Context().Value("photo").(Photo),
		Success: true,
	}
	WriteJsonResponse(v, w)
}

func (s *Server) GetPhoto(w http.ResponseWriter, r *http.Request) {
	s.GetImage("%s.jpg", w, r)
}

func (s *Server) GetThumbnail(w http.ResponseWriter, r *http.Request) {
	s.GetImage("%s-thumb.jpg", w, r)
}

func (s *Server) GetImage(imageFormat string, w http.ResponseWriter, r *http.Request) {
	photo := r.Context().Value("photo").(Photo)
	photoPath := path.Join(s.cfg.ImgFolder(), fmt.Sprintf(imageFormat, photo.ID))
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
