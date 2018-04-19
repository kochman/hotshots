package pusher

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/kochman/hotshots/config"
	"github.com/kochman/hotshots/log"
)

// Pusher is responsible for uploading photos from a camera to a remote location.
// It only uploads photos that don't exist remotely, in an effort to reduce bandwidth usage.
type Pusher struct {
	cfg               config.Config
	cameraService     cameraService
	filenameToPhotoID map[string]string
	photoService      photoService
}

// New creates a new Pusher.
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

// Run runs the Pusher's upload functionality in a loop forever.
func (p *Pusher) Run() {
	ticker := time.NewTicker(p.cfg.RefreshInterval)
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

	if len(toUpload) > 0 {
		log.Infof("%d existing photos on server, %d to upload", len(photos), len(toUpload))
	}

	for _, filename := range toUpload {
		b, err := p.cameraService.getFile(filename)
		if err != nil {
			log.WithError(err).Error("unable to get file")
			continue
		}

		log.Infof("uploading photo %s", filename)
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

	for _, filename := range filenames {
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
