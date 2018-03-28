package pusher

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kochman/hotshots/config"
)

type mockCamera struct {
	mock.Mock
}

func (mc *mockCamera) getFile(file string) ([]byte, error) {
	args := mc.Called(file)
	return args.Get(0).([]byte), args.Error(1)
}

func (mc *mockCamera) listFilenames() ([]string, error) {
	args := mc.Called()
	return args.Get(0).([]string), args.Error(1)
}

// type camera interface {
// 	listFilenames() ([]string, error)
// 	getFile(filename string) ([]byte, error)
// }

func TestGeneratePhotoID(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	p, err := New(cfg)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	input := []byte("hello there")
	output, err := p.generatePhotoID(input)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if output != "6e71b3cac15d32fe2d36c270887df9479c25c640" {
		t.Error("unexpected output")
	}
}

func TestGeneratePhotoIDsNotJPG(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	p, err := New(cfg)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	camera := &mockCamera{}
	camera.On("listFilenames").Return([]string{"hi.NEF", "there.CR2"}, nil)
	p.camera = camera

	p.generatePhotoIDs()

	camera.AssertExpectations(t)
}

func TestGeneratePhotoIDs(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	p, err := New(cfg)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	camera := &mockCamera{}
	camera.On("listFilenames").Return([]string{"hi.JPG", "there.JPG"}, nil)
	camera.On("getFile", "hi.JPG").Return([]byte("bloop"), nil)
	camera.On("getFile", "there.JPG").Return([]byte("bloop"), nil)
	p.camera = camera

	p.generatePhotoIDs()

	camera.AssertExpectations(t)
}

// func TestGetFile(t *testing.T) {
//
// }
