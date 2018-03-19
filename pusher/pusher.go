package pusher

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/kochman/hotshots/config"
	"github.com/kochman/hotshots/log"
)

const photosEndpoint = "/photos"
const photoIDEndpoint = photosEndpoint + "/ids"

// this is copied/pasted from the server package... should we just import it?
type photoIDsResponse struct {
	Success bool     `json:"success"`
	IDs     []string `json:"ids"`
}

// same as above... this is copied/pasted from the server package... should we just import it?
type photoResponse struct {
	Success bool `json:"success"`
	// Status  Status `json:"status"`
	NewID string `json:"new_id"`
}

type Pusher struct {
	cfg    config.Config
	camera camera
}

func New(cfg *config.Config) (*Pusher, error) {
	p := &Pusher{
		cfg:    *cfg,
		camera: newGphotoCamera(),
	}

	return p, nil
}

func (p *Pusher) Run() {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		photos, err := p.getExistingPhotos()
		if err != nil {
			log.WithError(err).Error("unable to get existing photos")
			continue
		}
		log.Infof("%+v", photos)

		filenames, err := p.camera.listFilenames()
		if err == errCameraNotConnected {
			log.Info("camera not connected")
			continue
		} else if err != nil {
			log.WithError(err).Error("unable to get filenames")
			continue
		}

		log.Infof("%+v", filenames)
		for _, filename := range filenames {
			// only transfer JPEGs
			if strings.ToLower(path.Ext(filename)) != ".jpg" {
				continue
			}
			b, err := p.camera.getFile(filename)
			if err != nil {
				log.WithError(err).Error("unable to get file")
				continue
			}

			err = p.uploadPhoto(b)
			if err != nil {
				log.WithError(err).Errorf("unable to upload %s", filename)
				continue
			}
			log.Infof("uploaded photo %s", filename)
		}
	}
}

func (p *Pusher) getExistingPhotos() ([]string, error) {
	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", p.cfg.PusherDestURL+photoIDEndpoint, nil)
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

func (p *Pusher) uploadPhoto(photo []byte) error {
	c := &http.Client{
		Timeout: 30 * time.Second,
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

	req, err := http.NewRequest("POST", p.cfg.PusherDestURL+photosEndpoint, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
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
