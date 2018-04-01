package pusher

import (
	"bytes"
	"errors"
	"io"
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
	unknownModel = "Unknown model"
)

// cameraService provides methods for transferring data from a camera
type cameraService interface {
	listFilenames() ([]string, error)
	getFile(filename string) ([]byte, error)
}

// camera wraps gphoto2go.Camera in order to make this package easier to test.
type gphoto2Camera interface {
	Init() int
	Exit() int
	RListFolders(folder string) []string
	ListFiles(folder string) ([]string, int)
	FileReader(folder, name string) io.ReadCloser
}

// localCamera communicates with a local camera through libgphoto2.
// All operations that interact with the camera are surrounded with a mutex.
type localCamera struct {
	m       *sync.Mutex
	gphoto2 gphoto2Camera
}

// newLocalCamera creates a new localCamera with internal mutex.
func newLocalCamera() *localCamera {
	return &localCamera{
		m:       &sync.Mutex{},
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
	defer c.exitCamera()

	folders := c.gphoto2.RListFolders("/")
	for _, folder := range folders {
		files, err := c.gphoto2.ListFiles(folder)
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

func (c *localCamera) getFile(filename string) ([]byte, error) {
	folder := path.Dir(filename)
	name := path.Base(filename)

	c.m.Lock()
	defer c.m.Unlock()

	err := c.initCamera()
	if err != nil {
		return []byte{}, err
	}
	defer c.exitCamera()

	cameraFileReader := c.gphoto2.FileReader(folder, name)
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
