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

type mockPhotoService struct {
	mock.Mock
}

func (mps *mockPhotoService) existingPhotos() ([]string, error) {
	args := mps.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (mps *mockPhotoService) uploadPhoto(photo []byte) error {
	args := mps.Called(photo)
	return args.Error(0)
}

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

	_, ok := p.filenameToPhotoID["hi.NEF"]
	if ok {
		t.Errorf("did not expect \"hi.NEF\" to have a photo ID")
	}

	_, ok = p.filenameToPhotoID["hi.CR2"]
	if ok {
		t.Errorf("did not expect \"hi.CR2\" to have a photo ID")
	}

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
	camera.On("getFile", "hi.JPG").Return([]byte("hello"), nil)
	camera.On("getFile", "there.JPG").Return([]byte("there"), nil)
	p.camera = camera

	p.generatePhotoIDs()

	val, ok := p.filenameToPhotoID["hi.JPG"]
	if ok {
		if val != "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d" {
			t.Errorf("\"hi.JPG\" has unexpected photo ID")
		}
	} else {
		t.Errorf("expected \"hi.JPG\" to have a photo ID")
	}

	val, ok = p.filenameToPhotoID["there.JPG"]
	if ok {
		if val != "490528f36debf7c15cea5e9a9d1ea024cf6b2921" {
			t.Errorf("\"there.JPG\" has unexpected photo ID")
		}
	} else {
		t.Errorf("expected \"there.JPG\" to have a photo ID")
	}

	camera.AssertExpectations(t)
}

func TestUploadNewPhotos(t *testing.T) {
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
	camera.On("getFile", "hi.JPG").Return([]byte("hello"), nil)
	camera.On("getFile", "there.JPG").Return([]byte("there"), nil)
	p.camera = camera

	photoService := &mockPhotoService{}
	photoService.On("existingPhotos").Return([]string{}, nil)
	photoService.On("uploadPhoto", mock.Anything).Return(nil)
	p.photoService = photoService

	p.uploadNewPhotos()

	camera.AssertExpectations(t)

	photoService.AssertExpectations(t)
	photoService.AssertNumberOfCalls(t, "existingPhotos", 1)
	photoService.AssertNumberOfCalls(t, "uploadPhoto", 2)
}
