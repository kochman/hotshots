package pusher

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/kochman/hotshots/config"
	"github.com/kochman/hotshots/log"
)

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
	cfg               config.Config
	cameraService     cameraService
	filenameToPhotoID map[string]string
	photoService      photoService
}

func New(cfg *config.Config) (*Pusher, error) {
	p := &Pusher{
		cfg:               *cfg,
		cameraService:     newLocalCamera(),
		filenameToPhotoID: map[string]string{},
		photoService: &remoteAPI{
			url: cfg.ServerURL,
		},
	}

	return p, nil
}

func (p *Pusher) Run() {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		p.uploadNewPhotos()
	}
}

func (p *Pusher) uploadNewPhotos() {
	p.generatePhotoIDs()

	photos, err := p.photoService.existingPhotos()
	if err != nil {
		log.WithError(err).Error("unable to get existing photos")
		return
	}
	log.Infof("%+v", photos)

	toUpload := []string{}
	for filename, id := range p.filenameToPhotoID {
		found := false
		for _, existing := range photos {
			if id == existing {
				found = true
				break
			}
		}
		if !found {
			toUpload = append(toUpload, filename)
		}
	}

	log.Infof("%+v", toUpload)
	for _, filename := range toUpload {
		// only transfer JPEGs
		if strings.ToLower(path.Ext(filename)) != ".jpg" {
			continue
		}

		b, err := p.cameraService.getFile(filename)
		if err != nil {
			log.WithError(err).Error("unable to get file")
			continue
		}

		err = p.photoService.uploadPhoto(b)
		if err != nil {
			log.WithError(err).Errorf("unable to upload %s", filename)
			continue
		}
		log.Infof("uploaded photo %s", filename)
	}

}

func (p *Pusher) generatePhotoID(photo []byte) (string, error) {
	digest := sha1.New()
	photoBuf := bytes.NewBuffer(photo)
	if _, err := photoBuf.WriteTo(digest); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}

func (p *Pusher) generatePhotoIDs() {
	filenames, err := p.cameraService.listFilenames()
	if err == errCameraNotConnected {
		return
	} else if err != nil {
		log.WithError(err).Error("unable to get filenames")
		return
	}

	log.Infof("%+v", filenames)
	for _, filename := range filenames {
		// only transfer JPEGs
		if strings.ToLower(path.Ext(filename)) != ".jpg" {
			continue
		}

		// generate an ID for this photo if not exists
		if _, ok := p.filenameToPhotoID[filename]; ok {
			continue
		}
		b, err := p.cameraService.getFile(filename)
		if err != nil {
			log.WithError(err).Error("unable to get file")
			continue
		}
		id, err := p.generatePhotoID(b)
		if err != nil {
			log.WithError(err).Errorf("unable to generate photo ID %s", filename)
			continue
		}

		p.filenameToPhotoID[filename] = id
	}
}
