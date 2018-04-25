package pusher

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/kochman/hotshots/server"
)

type photoService interface {
	existingPhotos() ([]string, error)
	uploadPhoto([]byte) error
}

type remoteAPI struct {
	url           string
	uploadTimeout time.Duration
}

const photosEndpoint = "/photos"
const photoIDEndpoint = photosEndpoint + "/ids"

var errPhotoExists = errors.New("photo already exists")

// existingPhotos returns all photo IDs that the remote server currently knows about.
func (r *remoteAPI) existingPhotos() ([]string, error) {
	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	ids := []string{}
	start := 0
	limit := 20

	for {
		req, err := http.NewRequest("GET", r.url+photoIDEndpoint, nil)
		if err != nil {
			return []string{}, err
		}

		query := req.URL.Query()
		query.Set("start", strconv.Itoa(start))
		query.Set("limit", strconv.Itoa(limit))
		req.URL.RawQuery = query.Encode()

		resp, err := c.Do(req)
		if err != nil {
			return []string{}, err
		}

		var idsResp server.GetPhotoIDsResponse
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&idsResp)
		if err != nil {
			return []string{}, err
		}

		ids = append(ids, idsResp.IDs...)

		if len(idsResp.IDs) < 20 {
			break
		}
		start += limit
	}

	return ids, nil
}

// uploadPhoto uploads a photo to the remote server.
func (r *remoteAPI) uploadPhoto(photo []byte) error {
	c := &http.Client{
		Timeout: r.uploadTimeout,
	}

	buf := bytes.Buffer{}
	writer := multipart.NewWriter(&buf)
	w, err := writer.CreateFormFile("photo", "photo")
	if err != nil {
		return err
	}

	photoBuf := bytes.NewBuffer(photo)
	_, err = photoBuf.WriteTo(w)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", r.url+photosEndpoint, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 400 {
		return errPhotoExists
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var photoResp server.PostPhotoResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&photoResp)
	if err != nil {
		return err
	}

	if !photoResp.Success {
		return errors.New("unsuccessful response")
	}

	return nil
}
