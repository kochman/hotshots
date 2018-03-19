package server

import (
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const emptyFilename = "_testdata/emptyFile.jpg"
const emptyID = "da39a3ee5e6b4b0d3255bfef95601890afd80709"

const garbageFilename = "_testdata/garbageFile.jpg"
const garbageID = "273a20e7e44fe68d24076ceca0843462ac0c2e2b"

const exiflessFilename = "_testdata/exiflessFile.jpg"
const exiflessID = "a78999ffcc074d8d05e1ff31707e5221373979e7"

const validFilename = "_testdata/validFile.jpg"
const validID = "064c350707017163f692032adb2db6cdddc6fab0"

type ThumbParam struct {
	largeWidth  int
	largeHeight int
	thumbWidth  int
	thumbHeight int
}

type MockWaitGroup struct {
	mock.Mock
}

func (wg *MockWaitGroup) Add(delta int) {
	wg.Called(delta)
}

func (wg *MockWaitGroup) Done() {
	wg.Called()
}

func (wg *MockWaitGroup) Wait() {
	wg.Called()
}

func testGenPhotoIDHelper(t *testing.T, filename, expectedID string) {
	f, err := os.Open(filename)
	require.Nil(t, err)
	defer f.Close()

	id, err := GenPhotoID(f)

	assert.Nil(t, err)
	assert.Equal(t, expectedID, id)
	pos, _ := f.Seek(0, io.SeekCurrent)
	assert.EqualValues(t, pos, 0)
}

func TestGenPhotoID(t *testing.T) {
	testGenPhotoIDHelper(t, emptyFilename, emptyID)
	testGenPhotoIDHelper(t, garbageFilename, garbageID)
	testGenPhotoIDHelper(t, exiflessFilename, exiflessID)
	testGenPhotoIDHelper(t, validFilename, validID)

	_, err := GenPhotoID(nil)
	assert.NotNil(t, err)
}

func testGetExifHelper(t *testing.T, filename string, valid bool) {
	imageFile, err := os.Open(filename)
	require.Nil(t, err)
	defer imageFile.Close()

	xif := make(chan *exif.Exif, 1)
	errs := make(chan error, 1)

	wg := new(MockWaitGroup)
	wg.On("Done").Return()

	GetExif(xif, imageFile, wg, errs)

	wg.AssertNumberOfCalls(t, "Done", 1)

	if valid {
		select {
		case x := <-xif: // exif sent as expected
			assert.NotNil(t, x)
		default: // no exif sent
			t.Error("exif not sent to go channel")
		}
	} else {
		select {
		case err := <-errs: // exif sent as expected
			assert.NotNil(t, err)
		default: // no exif sent
			t.Error("err not sent to go channel")
		}
	}
}

func TestGetExif(t *testing.T) {
	testGetExifHelper(t, emptyFilename, false)
	testGetExifHelper(t, garbageFilename, false)
	testGetExifHelper(t, exiflessFilename, false)
	testGetExifHelper(t, validFilename, true)
}

func testSaveThumbHelper(t *testing.T, filename string, valid bool, param *ThumbParam) {
	imageFile, err := os.Open(filename)
	require.Nil(t, err)
	defer imageFile.Close()

	thumbFile, err := ioutil.TempFile("", "thumb")
	require.Nil(t, err)
	defer func() {
		thumbFile.Close()
		os.Remove(thumbFile.Name())
	}()

	rect := make(chan *image.Rectangle, 1)
	errs := make(chan error, 1)

	wg := new(MockWaitGroup)
	wg.On("Done").Return()

	SaveThumb(rect, imageFile, thumbFile.Name(), wg, errs)

	wg.AssertNumberOfCalls(t, "Done", 1)

	if valid {
		select {
		case r := <-rect: // exif sent as expected
			require.NotNil(t, r)
			assert.EqualValues(t, r.Dx(), param.largeWidth)
			assert.EqualValues(t, r.Dy(), param.largeHeight)
		default: // no exif sent
			t.Error("exif not sent to go channel")
		}
		i, err := jpeg.Decode(thumbFile)
		require.Nil(t, err)
		assert.EqualValues(t, i.Bounds().Dx(), param.thumbWidth)
		assert.EqualValues(t, i.Bounds().Dy(), param.thumbHeight)
	} else {
		select {
		case err := <-errs: // exif sent as expected
			assert.NotNil(t, err)
		default: // no exif sent
			t.Error("err not sent to go channel")
		}
	}
}

func TestSaveThumb(t *testing.T) {
	testSaveThumbHelper(t, emptyFilename, false, nil)
	testSaveThumbHelper(t, garbageFilename, false, nil)
	testSaveThumbHelper(t, exiflessFilename, true, &ThumbParam{6000, 4000, 256, 170})
	testSaveThumbHelper(t, validFilename, true, &ThumbParam{6000, 4000, 256, 170})
}

func testSaveImageHelper(t *testing.T, filename string, valid bool) {
	inputFile, err := os.Open(filename)
	require.Nil(t, err)
	defer inputFile.Close()

	outputFile, err := ioutil.TempFile("", "output")
	require.Nil(t, err)
	defer func() {
		outputFile.Close()
		os.Remove(outputFile.Name())
	}()

	errs := make(chan error, 1)

	wg := new(MockWaitGroup)
	wg.On("Done").Return()

	SaveImage(inputFile, outputFile.Name(), wg, errs)

	wg.AssertNumberOfCalls(t, "Done", 1)

	if valid {
		inputFile.Seek(0, io.SeekStart)
		inputID, err := GenPhotoID(inputFile)
		require.Nil(t, err)
		outputID, err := GenPhotoID(outputFile)
		require.Nil(t, err)
		assert.EqualValues(t, inputID, outputID)
	} else {
		select {
		case err := <-errs: // exif sent as expected
			assert.NotNil(t, err)
		default: // no exif sent
			t.Error("err not sent to go channel")
		}
	}
}

func TestSaveImage(t *testing.T) {
	testSaveImageHelper(t, emptyFilename, false)
	testSaveImageHelper(t, garbageFilename, false)
	testSaveImageHelper(t, exiflessFilename, true)
	testSaveImageHelper(t, validFilename, true)
}

func testProcessPhotoHelper(t *testing.T, filename, expectedID string, valid bool) {
	input, err := os.Open(filename)
	require.Nil(t, err)
	defer input.Close()

	outputFile, err := ioutil.TempFile("", "output")
	require.Nil(t, err)
	defer func() {
		outputFile.Close()
		os.Remove(outputFile.Name())
	}()

	thumbFile, err := ioutil.TempFile("", "thumb")
	require.Nil(t, err)
	defer func() {
		thumbFile.Close()
		os.Remove(thumbFile.Name())
	}()

	start := time.Now()
	xif, rect, err := ProcessPhoto(input, expectedID, outputFile.Name(), thumbFile.Name(), time.Minute*2)
	stop := time.Now()

	require.WithinDuration(t, start, stop, 2*time.Minute, "time to execute photo processing")

	if valid {
		assert.Nil(t, err)
		require.NotNil(t, xif)
		require.NotNil(t, rect)
		outputInfo, err := os.Stat(outputFile.Name())
		require.Nil(t, err)
		thumbInfo, err := os.Stat(thumbFile.Name())
		require.Nil(t, err)
		assert.NotZero(t, outputInfo.Size())
		assert.NotZero(t, thumbInfo.Size())

	} else {
		require.NotNil(t, err)
		assert.Nil(t, xif)
		assert.Nil(t, rect)
	}
}

func TestProcessPhoto(t *testing.T) {
	testProcessPhotoHelper(t, emptyFilename, emptyID, false)
	testProcessPhotoHelper(t, garbageFilename, garbageID, false)
	testProcessPhotoHelper(t, exiflessFilename, exiflessID, false)
	testProcessPhotoHelper(t, validFilename, validID, true)
}
