package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/asdine/storm"
	"github.com/go-chi/chi"
	"github.com/kochman/hotshots/log"
)

/*
 * Response Structs
 */

type GetPhotoIDsResponse struct {
	Success bool     `json:"success"`
	IDs     []string `json:"ids"`
}

type GetPhotosResponse struct {
	Success bool    `json:"success"`
	Photos  []Photo `json:"photos"`
}

type PostPhotoResponse struct {
	Success bool   `json:"success"`
	Status  Status `json:"status"`
	NewID   string `json:"new_id"`
}

type GetPhotoMetadataResponse struct {
	Success bool  `json:"success"`
	Photo   Photo `json:"photo"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type NotFoundResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	URI     string `json:"uri"`
}

/*
 * Handlers
 */

func NotFound(w http.ResponseWriter, r *http.Request) {
	v := NotFoundResponse{
		Success: false,
		Error:   "requested URI not found",
		URI:     r.RequestURI,
	}
	WriteJsonResponse(v, 404, w)
}

func (s *Server) PhotoCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		photoID := chi.URLParam(r, "pid")
		photo, err := s.GetPhotoFromDatabase(photoID)
		if err != nil {
			log.Info(err)
			WriteError("unable to find photo id", 404, w)
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
	if err := s.db.Find("Status", ProcessingSucceeded, &photos, start, limit); err != nil && err != storm.ErrNotFound {
		log.Error(err)
		WriteError("unable to query photos", 500, w)
		return
	}

	v := GetPhotosResponse{
		Success: true,
		Photos:  photos,
	}
	WriteJsonResponse(v, 200, w)
}

func (s *Server) GetPhotoIDs(w http.ResponseWriter, r *http.Request) {
	start, limit, err := GetPaginateValues(r)
	if err != nil {
		log.Error(err)
		WriteError("Unable to parse query string", 400, w)
		return
	}
	var photos []Photo
	if err := s.db.Find("Status", ProcessingSucceeded, &photos, start, limit); err != nil && err != storm.ErrNotFound {
		log.Error(err)
		WriteError("unable to query photos", 500, w)
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
	WriteJsonResponse(v, 200, w)
}

func (s *Server) PostPhoto(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(1 << 25) // 32M max memory

	input, _, err := r.FormFile("photo")
	if err != nil {
		log.Info(err)
		WriteError("unable to parse form value 'photo'", 400, w)
		return
	}
	defer input.Close()

	id, err := GenPhotoID(input)
	if err != nil {
		log.Error(err)
		WriteError("unable to generate photo ID", 500, w)
		return
	}

	overwrite := r.FormValue("overwrite")
	photo, err := s.GetPhotoFromDatabase(id)
	if err == nil {
		if overwrite == "true" {
			s.db.DeleteStruct(photo)
		} else {
			WriteError("photo already exists", 400, w)
			return
		}
	} else if err != storm.ErrNotFound {
		log.Error(err)
		WriteError("unable to read from database", 500, w)
		return
	}

	now := time.Now()
	photo = Photo{
		ID:              id,
		UploadedAt:      &now,
		Status:          Processing,
		StatusUpdatedAt: &now,
	}

	if err := s.db.Save(&photo); err != nil {
		log.Error(err)
		WriteError("unable to write to database", 500, w)
		return
	}
	v := PostPhotoResponse{
		Success: true,
		NewID:   id,
		Status:  Processing,
	}
	WriteJsonResponse(v, 200, w)

	go func() {
		photoPath := path.Join(s.cfg.ImgFolder(), fmt.Sprintf("%s.jpg", id))
		thumbPath := path.Join(s.cfg.ImgFolder(), fmt.Sprintf("%s-thumb.jpg", id))

		xif, rect, err := ProcessPhoto(input, id, photoPath, thumbPath, s.timeout)
		if err != nil {
			log.Error(err)
			photo.UpdateStatus(ProcessingFailed)
			if err := s.db.Update(&photo); err != nil {
				log.Error(err)
			}

			os.Remove(photoPath)
			os.Remove(thumbPath)
			return
		}

		photo.AddMetadata(rect, xif)
		photo.UpdateStatus(ProcessingSucceeded)

		if err := s.db.Update(&photo); err != nil {
			log.Error(err)

			os.Remove(photoPath)
			os.Remove(thumbPath)
			return
		}
	}()

}

func (s *Server) GetPhotoMetadata(w http.ResponseWriter, r *http.Request) {
	v := GetPhotoMetadataResponse{
		Photo:   r.Context().Value("photo").(Photo),
		Success: true,
	}
	WriteJsonResponse(v, 200, w)
}

func (s *Server) GetPhoto(w http.ResponseWriter, r *http.Request) {
	s.GetImage("%s.jpg", w, r)
}

func (s *Server) GetThumbnail(w http.ResponseWriter, r *http.Request) {
	s.GetImage("%s-thumb.jpg", w, r)
}

func (s *Server) GetImage(imageFormat string, w http.ResponseWriter, r *http.Request) {
	photo := r.Context().Value("photo").(Photo)
	if photo.Status != ProcessingSucceeded {
		WriteError("photo not processed", 400, w)
		return
	}
	photoPath := path.Join(s.cfg.ImgFolder(), fmt.Sprintf(imageFormat, photo.ID))
	output, err := os.OpenFile(photoPath, os.O_RDONLY, 0660)
	if err != nil {
		log.Error(err)
		WriteError("unable to access internal storage of image", 500, w)
		return
	}
	defer output.Close()

	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(200)
	if _, err := io.Copy(w, output); err != nil {
		log.Error(err)
		WriteError("unable to return image data", 500, w)
		return
	}
}
