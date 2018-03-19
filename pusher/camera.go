package pusher

import (
	"bytes"
	"errors"
	"path"
	"sync"

	"github.com/micahwedemeyer/gphoto2go"
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
	libgphoto2UnknownModel = "Unknown model"
)

// camera wraps gphoto2go's Camera in order to make this package easier to test.
type camera interface {
	listFilenames() ([]string, error)
	getFile(filename string) ([]byte, error)
}

// All operations that interact with the camera are surrounded with a mutex.
type gphotoCamera struct {
	m      *sync.Mutex
	camera gphoto2go.Camera
}

func newGphotoCamera() *gphotoCamera {
	return &gphotoCamera{
		m:      &sync.Mutex{},
		camera: gphoto2go.Camera{},
	}
}

func (gc *gphotoCamera) initCamera() error {
	err := gc.camera.Init()
	if err < 0 {
		gphotoErr := gphoto2go.CameraResultToString(err)
		if gphotoErr == libgphoto2UnknownModel {
			return errCameraNotConnected
		}
		return &UnhandledError{msg: gphotoErr}
	}
	return nil
}

func (gc *gphotoCamera) exitCamera() error {
	err := gc.camera.Exit()
	if err < 0 {
		gphotoErr := gphoto2go.CameraResultToString(err)
		return &UnhandledError{msg: gphotoErr}
	}
	return nil
}

func (gc *gphotoCamera) listFilenames() ([]string, error) {
	filenames := []string{}

	err := gc.initCamera()
	if err != nil {
		return filenames, err
	}
	defer gc.exitCamera()

	folders := gc.camera.RListFolders("/")
	for _, folder := range folders {
		files, err := gc.camera.ListFiles(folder)
		if err < 0 {
			gphotoErr := gphoto2go.CameraResultToString(err)
			return filenames, &UnhandledError{msg: gphotoErr}
		}
		for _, file := range files {
			filename := path.Join(folder, file)
			filenames = append(filenames, filename)
		}
	}

	return filenames, nil
}

func (gc *gphotoCamera) getFile(filename string) ([]byte, error) {
	folder := path.Dir(filename)
	name := path.Base(filename)

	gc.m.Lock()
	defer gc.m.Unlock()

	err := gc.initCamera()
	if err != nil {
		return []byte{}, err
	}
	defer gc.exitCamera()

	cameraFileReader := gc.camera.FileReader(folder, name)
	// Important, since there is memory used in the transfer that needs to be freed up
	defer cameraFileReader.Close()

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
