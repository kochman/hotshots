package pusher

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"
)

type photoService interface {
	existingPhotos() ([]string, error)
	uploadPhoto([]byte) error
}

type remoteAPI struct {
	url string
}

const photosEndpoint = "/photos"
const photoIDEndpoint = photosEndpoint + "/ids"

var errPhotoExists = errors.New("photo already exists")

func (r *remoteAPI) existingPhotos() ([]string, error) {
	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", r.url+photoIDEndpoint, nil)
	if err != nil {
		return []string{}, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return []string{}, err
	}

	var idsResp photoIDsResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&idsResp)
	if err != nil {
		return []string{}, err
	}

	return idsResp.IDs, nil
}

func (r *remoteAPI) uploadPhoto(photo []byte) error {
	c := &http.Client{
		Timeout: 10 * time.Second,
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

	var photoResp photoResponse
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
