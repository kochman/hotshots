package server

import (
	"crypto/sha1"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/nfnt/resize"
	"github.com/kochman/hotshots/log"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

/*
 * Constants
 */

const MaxWidth = 256
const MaxHeight = 256

/*
 * Data structs
 */

type Photo struct {
	ID			string    `storm:"id"`
	UploadedAt	time.Time `storm:"index"`
	TakenAt		time.Time `storm:"index"`
	Width		int       `storm:"index"`
	Height		int       `storm:"index"`
	Megapixels	float64   `storm:"index"`
	Lat			float64
	Long		float64
	CamSerial	string    `storm:"index"`
	CamMake		string    `storm:"index"`
	CamModel	string    `storm:"index"`
}


func NewPhoto(id string, r image.Rectangle, x exif.Exif) *Photo{
	taken, err := x.DateTime()
	if err != nil {
		log.Error("Unable to find time for ", id)
	}

	lat, long, err := x.LatLong()
	if err != nil {
		log.Info("Unable to find gps for ", id)
	}

	cserial, err := x.Get(mknote.SerialNumber)
	if err != nil {
		log.Info("Unable to find serial for ", id)
	}

	cmake, err := x.Get(exif.Make)
	if err != nil {
		log.Info("Unable to find make for ", id)
	}

	cmodel, err := x.Get(exif.Model)
	if err != nil {
		log.Info("Unable to find model for ", id)
	}

	mp := toFixed(float64(r.Dx()) * float64(r.Dy()) / 1000000.0, 1)

	return &Photo{
		ID:         id,
		UploadedAt: time.Now(),
		TakenAt:	taken,
		Width:		r.Dx(),
		Height:		r.Dy(),
		Megapixels: mp,
		Lat:		lat,
		Long:		long,
		CamSerial:	strings.TrimSuffix(string(cserial.Val), "\u0001"),
		CamMake:	strings.TrimSuffix(string(cmake.Val), "\u0001"),
		CamModel:	strings.TrimSuffix(string(cmodel.Val), "\u0001"),
	}
}

func SaveImage(data io.Reader, path string, saveErr chan error) {
	output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		log.Error(err)
		saveErr <- err
		return
	}
	defer output.Close()

	if _, err := io.Copy(output, data); err != nil {
		log.Error(err)
		saveErr <- err
		return
	}

	saveErr <- nil
}

func SaveThumb(r *image.Rectangle, data io.Reader, path string, saveErr chan error) {
	img, err := jpeg.Decode(data)
	io.Copy(ioutil.Discard, data)
	if err != nil {
		log.Error(err)
		saveErr <- err
		return
	}
	*r = img.Bounds()
	resizedImg := resize.Thumbnail(MaxWidth, MaxHeight, img, resize.Bicubic)
	output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		log.Error(err)
		saveErr <- err
		return
	}
	defer output.Close()

	jpeg.Encode(output, resizedImg, nil)

	saveErr <- nil
}

func GetExif(out *exif.Exif, data io.Reader, saveErr chan error) {
	x, err := exif.Decode(data)
	io.Copy(ioutil.Discard, data)
	if err != nil {
		log.Error(err)
		saveErr <- err
		return
	}
	*out = *x

	saveErr <- nil
}


func GenPhotoID(f io.ReadSeeker) (string, error) {
	digest := sha1.New()
	if _, err := io.Copy(digest, f); err != nil {
		return "", err
	}
	f.Seek(0, io.SeekStart)
	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}