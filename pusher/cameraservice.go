package pusher

import (
	"bytes"
	"errors"
	"io"
	"path"
	"strings"

	"github.com/micahwedemeyer/gphoto2go"

	"github.com/kochman/hotshots/log"
)

var (
	errCameraNotConnected = errors.New("camera not connected")
	errFileNotTransferred = errors.New("file not transferred")
)

// UnhandledError implements the error interface, allowing us to wrap lower-level
// errors and return a useful representation to clients.
type UnhandledError struct {
	msg string
}

func (e *UnhandledError) Error() string {
	return "unhandled error: " + e.msg
}

// For converting libgphoto2 error strings into Go errors.
// See https://github.com/gphoto/libgphoto2/blob/libgphoto2-2_5_16-release/libgphoto2/gphoto2-result.c
const (
	unknownModel = "Unknown model"
)

// cameraService provides methods for transferring data from a camera
type cameraService interface {
	listFilenames() ([]string, error)
	getFile(filename string) ([]byte, error)
}

// gphoto2Camera is based on gphoto2go.Camera in order to make this package easier to test.
type gphoto2Camera interface {
	Init() int
	Exit() int
	RListFolders(folder string) []string
	ListFiles(folder string) ([]string, int)
	FileReader(folder, name string) io.ReadCloser
}

// localCamera communicates with a local camera through libgphoto2.
type localCamera struct {
	gphoto2 gphoto2Camera
}

// newLocalCamera creates a new localCamera with internal mutex.
func newLocalCamera() *localCamera {
	return &localCamera{
		gphoto2: &gphoto2go.Camera{},
	}
}

func (c *localCamera) initCamera() error {
	err := c.gphoto2.Init()
	if err < 0 {
		gphotoErr := gphoto2go.CameraResultToString(err)
		if gphotoErr == unknownModel {
			return errCameraNotConnected
		}
		return &UnhandledError{msg: gphotoErr}
	}
	return nil
}

func (c *localCamera) exitCamera() error {
	err := c.gphoto2.Exit()
	if err < 0 {
		gphotoErr := gphoto2go.CameraResultToString(err)
		return &UnhandledError{msg: gphotoErr}
	}
	return nil
}

func (c *localCamera) listFilenames() ([]string, error) {
	filenames := []string{}

	err := c.initCamera()
	if err != nil {
		return filenames, err
	}
	defer func() {
		err = c.exitCamera()
		if err != nil {
			log.WithError(err).Error("unable to exit camera")
		}
	}()

	folders := c.gphoto2.RListFolders("/")
	for _, folder := range folders {
		files, err := c.gphoto2.ListFiles(folder)
		if err < 0 {
			gphotoErr := gphoto2go.CameraResultToString(err)
			return filenames, &UnhandledError{msg: gphotoErr}
		}
		for _, file := range files {
			// only transfer JPEGs
			if strings.ToLower(path.Ext(file)) != ".jpg" {
				continue
			}

			filename := path.Join(folder, file)
			filenames = append(filenames, filename)
		}
	}

	return filenames, nil
}

func (c *localCamera) getFile(filename string) ([]byte, error) {
	folder := path.Dir(filename)
	name := path.Base(filename)

	err := c.initCamera()
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		err = c.exitCamera()
		if err != nil {
			log.WithError(err).Error("unable to exit camera")
		}
	}()

	cameraFileReader := c.gphoto2.FileReader(folder, name)
	// Important, since there is memory used in the transfer that needs to be freed up
	defer func() {
		err = cameraFileReader.Close()
		if err != nil {
			log.WithError(err).Error("unable to close camera file reader")
		}
	}()

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(cameraFileReader)
	if err != nil {
		return []byte{}, err
	}
	if buf.Len() == 0 {
		return []byte{}, errFileNotTransferred
	}

	return buf.Bytes(), nil
}
