package pusher

import (
	"path"
	"strings"
	"time"

	"github.com/kochman/hotshots/config"
	"github.com/kochman/hotshots/log"
)

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
			} else {
				log.Infof("%s len %d", filename, len(b))
			}
		}
	}
}
