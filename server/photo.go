package server

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kochman/hotshots/log"
	"github.com/nfnt/resize"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"github.com/rwcarlsen/goexif/tiff"
)

/*
 * Constants
 */

const MaxWidth = 256
const MaxHeight = 256
const ProcessingTimeout = 5

type Status uint8

const (
	Processing          Status = iota
	ProcessingSucceeded Status = iota
	ProcessingFailed    Status = iota
)

/*
 * Data structs
 */

type ProcessorWaitGroup interface {
	Add(delta int)
	Done()
	Wait()
}

type ExifData interface {
	DateTime() (time.Time, error)
	LatLong() (float64, float64, error)
	Get(name exif.FieldName) (*tiff.Tag, error)
}

type Photo struct {
	ID              string     `storm:"id" json:"id"`
	UploadedAt      *time.Time `storm:"index" json:"uploaded_at"`
	TakenAt         *time.Time `storm:"index" json:"taken_at"`
	Width           int        `storm:"index" json:"width"`
	Height          int        `storm:"index" json:"height"`
	Megapixels      float64    `storm:"index" json:"megapixels"`
	Lat             float64    `json:"lat"`
	Long            float64    `json:"long"`
	CamSerial       string     `storm:"index" json:"cam_serial"`
	CamMake         string     `storm:"index" json:"cam_make"`
	CamModel        string     `storm:"index" json:"cam_model"`
	Status          Status     `storm:"index" json:"status"`
	StatusUpdatedAt *time.Time `storm:"index" json:"status_updated_at"`
}

func (p *Photo) AddMetadata(r *image.Rectangle, x ExifData) {
	taken, err := x.DateTime()
	if err != nil {
		log.Error("unable to find time for ", p.ID)
	} else {
		p.TakenAt = &taken
	}

	lat, long, err := x.LatLong()
	if err != nil {
		log.Info("unable to find gps for ", p.ID)
	} else {
		p.Lat = lat
		p.Long = long
	}

	cserial, err := x.Get(mknote.SerialNumber)
	if err != nil {
		log.Info("unable to find serial for ", p.ID)
	} else {
		p.CamSerial = strings.TrimSuffix(string(cserial.Val), "\u0000")
	}

	cmake, err := x.Get(exif.Make)
	if err != nil {
		log.Info("unable to find make for ", p.ID)
	} else {
		p.CamMake = strings.TrimSuffix(string(cmake.Val), "\u0000")
	}

	cmodel, err := x.Get(exif.Model)
	if err != nil {
		log.Info("unable to find model for ", p.ID)
	} else {
		p.CamModel = strings.TrimSuffix(string(cmodel.Val), "\u0000")
	}

	p.Width = r.Dx()
	p.Height = r.Dy()
	p.Megapixels = toFixed(float64(r.Dx())*float64(r.Dy())/1000000.0, 2)
}

func (p *Photo) UpdateStatus(status Status) {
	p.Status = status
	t := time.Now()
	p.StatusUpdatedAt = &t
}

func ProcessPhoto(input io.Reader, id string, photoPath string, thumbPath string, timeout time.Duration) (*exif.Exif, *image.Rectangle, error) {
	rect := make(chan *image.Rectangle, 1)
	xif := make(chan *exif.Exif, 1)
	errs := make(chan error, 3)

	photoBuff := new(bytes.Buffer)
	thumbBuff := new(bytes.Buffer)
	exifBuff := new(bytes.Buffer)

	func() {
		mw := io.MultiWriter(photoBuff, thumbBuff, exifBuff)

		if _, err := io.Copy(mw, input); err != nil {
			log.Error(err)
			return
		}
	}()

	log.Info(fmt.Sprint("saving image ", id))

	wg := new(sync.WaitGroup)
	wg.Add(3)

	go SaveImage(photoBuff, photoPath, wg, errs)
	go SaveThumb(rect, thumbBuff, thumbPath, wg, errs)
	go GetExif(xif, exifBuff, wg, errs)

	if waitTimeout(wg, timeout) {
		log.Error("image processing timed out")
		return nil, nil, errors.New("image processing timed out")
	}
	if len(errs) != 0 {
		for err := range errs {
			return nil, nil, err
		}
	}
	x := <-xif
	r := <-rect
	return x, r, nil
}

func SaveImage(data io.Reader, path string, wg ProcessorWaitGroup, errs chan error) {
	jpgSignature := []byte{0xFF, 0xD8, 0xFF}
	sigCheck := make([]byte, 3)
	if _, err := io.ReadFull(data, sigCheck); err != nil {
		log.Error(err)
		errs <- err
		wg.Done()
		return
	}

	if !bytes.Equal(jpgSignature, sigCheck) {
		err := errors.New("file signature is incorrect")
		log.Error(err)
		errs <- err
		wg.Done()
		return
	}

	data = io.MultiReader(bytes.NewReader(sigCheck), data)

	output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		log.Error(err)
		errs <- err
		wg.Done()
		return
	}
	defer output.Close()

	if _, err := io.Copy(output, data); err != nil {
		log.Error(err)
		errs <- err
		wg.Done()
		return
	}

	wg.Done()
}

func SaveThumb(r chan *image.Rectangle, data io.Reader, path string, wg ProcessorWaitGroup, errs chan error) {
	img, err := jpeg.Decode(data)
	if err != nil {
		log.Error(err)
		errs <- err
		wg.Done()
		return
	}
	rect := img.Bounds()
	r <- &rect
	resizedImg := resize.Thumbnail(MaxWidth, MaxHeight, img, resize.Bicubic)
	output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		log.Error(err)
		errs <- err
		wg.Done()
		return
	}
	defer output.Close()

	if err := jpeg.Encode(output, resizedImg, nil); err != nil {
		log.Error(err)
		errs <- err
	}

	wg.Done()
}

func GetExif(out chan *exif.Exif, data io.Reader, wg ProcessorWaitGroup, errs chan error) {
	x, err := exif.Decode(data)
	if err != nil {
		log.Error(err)
		errs <- err
		wg.Done()
		return
	}
	out <- x

	wg.Done()
}

func (s *Server) GetPhotoFromDatabase(photoID string) (Photo, error) {
	var photo Photo
	if err := s.db.One("ID", photoID, &photo); err != nil {
		return Photo{}, err
	}
	return photo, nil
}

func GenPhotoID(f io.ReadSeeker) (string, error) {
	if f == nil {
		return "", errors.New("f is nil")
	}

	digest := sha1.New()
	if _, err := io.Copy(digest, f); err != nil {
		return "", err
	}
	f.Seek(0, io.SeekStart)
	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}

func (s *Status) String() string {
	switch *s {
	case Processing:
		return "processing"
	case ProcessingSucceeded:
		return "processing succeeded"
	case ProcessingFailed:
		return "processing failed"
	default:
		return ""
	}
}

func ToStatus(s string) Status {
	switch s {
	case "processing":
		return Processing
	case "processing succeeded":
		return ProcessingSucceeded
	case "processing failed":
		return ProcessingFailed
	default:
		return 255
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return new(json.InvalidUnmarshalError)
	}
	*s = ToStatus(str)
	if *s == 255 {
		return new(json.InvalidUnmarshalError)
	}
	return nil
}
