package pusher

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/mock"
)

type mockGphoto2Camera struct {
	mock.Mock
}

func (m *mockGphoto2Camera) Init() int {
	args := m.Called()
	return args.Int(0)
}

func (m *mockGphoto2Camera) Exit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *mockGphoto2Camera) RListFolders(folder string) []string {
	args := m.Called(folder)
	return args.Get(0).([]string)
}

func (m *mockGphoto2Camera) ListFiles(folder string) ([]string, int) {
	args := m.Called(folder)
	return args.Get(0).([]string), args.Int(1)
}

func (m *mockGphoto2Camera) FileReader(folder, name string) io.ReadCloser {
	args := m.Called(folder, name)
	return args.Get(0).(io.ReadCloser)
}

func TestListFilenames(t *testing.T) {
	c := newLocalCamera()

	gphoto2Camera := &mockGphoto2Camera{}
	gphoto2Camera.On("Init").Return(0)
	gphoto2Camera.On("RListFolders", "/").Return([]string{"testdir", "testdir2"})
	gphoto2Camera.On("ListFiles", "testdir").Return([]string{"hello.JPG", "there.JPG"}, 0)
	gphoto2Camera.On("ListFiles", "testdir2").Return([]string{"haha.JPG"}, 0)
	gphoto2Camera.On("Exit").Return(0)
	c.gphoto2 = gphoto2Camera

	filenames, err := c.listFilenames()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if len(filenames) != 3 {
		t.Errorf("got %d filenames, expected 3", len(filenames))
		return
	}
	if filenames[0] != "testdir/hello.JPG" {
		t.Errorf("got unexpected filename: %s", filenames[0])
	}
	if filenames[1] != "testdir/there.JPG" {
		t.Errorf("got unexpected filename: %s", filenames[2])
	}
	if filenames[2] != "testdir2/haha.JPG" {
		t.Errorf("got unexpected filename: %s", filenames[3])
	}

	gphoto2Camera.AssertExpectations(t)
	gphoto2Camera.AssertNumberOfCalls(t, "Init", 1)
	gphoto2Camera.AssertNumberOfCalls(t, "Exit", 1)
	gphoto2Camera.AssertNumberOfCalls(t, "RListFolders", 1)
	gphoto2Camera.AssertNumberOfCalls(t, "ListFiles", 2)
}

func TestGetFile(t *testing.T) {
	c := newLocalCamera()

	gphoto2Camera := &mockGphoto2Camera{}
	gphoto2Camera.On("Init").Return(0)
	body := ioutil.NopCloser(bytes.NewBufferString("hello"))
	gphoto2Camera.On("FileReader", "testdir", "hello.JPG").Return(body)
	gphoto2Camera.On("Exit").Return(0)
	c.gphoto2 = gphoto2Camera

	file, err := c.getFile("testdir/hello.JPG")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if string(file) != "hello" {
		t.Errorf("got unexpected file contents")
	}

	gphoto2Camera.AssertExpectations(t)
	gphoto2Camera.AssertNumberOfCalls(t, "Init", 1)
	gphoto2Camera.AssertNumberOfCalls(t, "Exit", 1)
	gphoto2Camera.AssertNumberOfCalls(t, "FileReader", 1)
}
