package server

import (
	"encoding/json"
	"image"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"github.com/rwcarlsen/goexif/tiff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const emptyPhotoJSON = `{"id":"","deleted":false,"uploaded_at":null,"taken_at":null,"width":0,"height":0,"megapixels":0,"lat":0,"long":0,"cam_serial":"","cam_make":"","cam_model":"","status":"processing","status_updated_at":null,"tags":null}`

const emptyJSON = `{}`

type MockExif struct {
	mock.Mock
}

func (x *MockExif) DateTime() (time.Time, error) {
	args := x.Called()
	return args.Get(0).(time.Time), args.Error(1)
}

func (x *MockExif) LatLong() (float64, float64, error) {
	args := x.Called()
	return args.Get(0).(float64), args.Get(1).(float64), args.Error(2)
}

func (x *MockExif) Get(name exif.FieldName) (*tiff.Tag, error) {
	args := x.Called(name)
	return args.Get(0).(*tiff.Tag), args.Error(1)
}

func AreEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(o1, o2), nil
}

func TestEmptyPhotoMarshal(t *testing.T) {
	input := new(Photo)
	bytea, err := json.Marshal(input)
	if err != nil {
		t.Error("unable to marshal photo")
	}

	str := string(bytea[:])

	if eq, _ := AreEqualJSON(emptyPhotoJSON, str); !eq {
		t.Error(
			"\nexpected: ", emptyPhotoJSON,
			"\nactual:   ", str,
		)
	}
}

func TestEmptyPhotoUnmarshal(t *testing.T) {
	var photoA, photoB Photo
	if err := json.Unmarshal([]byte(emptyJSON), &photoA); err != nil {
		t.Error("unable to unmarshal photo")
	}
	if err := json.Unmarshal([]byte(emptyPhotoJSON), &photoA); err != nil {
		t.Error("unable to unmarshal photo")
	}

	var expected Photo // empty photo struct
	assert.Equal(t, expected, photoA)
	assert.Equal(t, expected, photoB)
}

func testPhotoMarshalStatusHelper(t *testing.T, s Status) {
	var photo Photo
	photo.Status = s
	bytea, err := json.Marshal(photo)
	if err != nil {
		t.Error("unable to marshal photo")
	}
	str := string(bytea[:])
	if !strings.Contains(str, "\"status\":\""+s.String()+"\"") {
		t.Error("unable to locate correct status serialization for ", s)
	}
}

func TestPhotoMarshalStatus(t *testing.T) {
	testPhotoMarshalStatusHelper(t, Processing)
	testPhotoMarshalStatusHelper(t, ProcessingSucceeded)
	testPhotoMarshalStatusHelper(t, ProcessingFailed)
	testPhotoMarshalStatusHelper(t, 123)
}

func testPhotoUnmarshalStatusHelper(t *testing.T, s string, valid bool) {
	str := strings.Replace(emptyPhotoJSON, "\"status\":\"processing\"", "\"status\":\""+s+"\"", 1)

	var photo Photo
	err := json.Unmarshal([]byte(str), &photo)

	if valid {
		if err != nil {
			t.Error("unable to unmarshal photo")
		}

		assert.Equal(t, s, photo.Status.String())
	} else {
		if err == nil {
			t.Error("should error on unmarshal parsing issue")
		}

		assert.Equal(t, "", photo.Status.String())
	}
}

func TestPhotoUnmarshalStatus(t *testing.T) {
	testPhotoUnmarshalStatusHelper(t, "processing", true)
	testPhotoUnmarshalStatusHelper(t, "processing succeeded", true)
	testPhotoUnmarshalStatusHelper(t, "processing failed", true)
	testPhotoUnmarshalStatusHelper(t, "fasfasd", false)
	testPhotoUnmarshalStatusHelper(t, "", false)
}

func TestPhotoUpdateStatus(t *testing.T) {
	var p Photo
	before := time.Now()
	p.StatusUpdatedAt = &before
	time.Sleep(time.Millisecond)
	p.UpdateStatus(ProcessingSucceeded)

	assert.Equal(t, p.Status, ProcessingSucceeded)
	assert.True(t, p.StatusUpdatedAt.After(before), "updated time is after the first time")
}

func TestPhotoAddMetadata(t *testing.T) {
	var p Photo
	var SerialNumber exif.FieldName = mknote.SerialNumber
	var Make exif.FieldName = exif.Make
	var Model exif.FieldName = exif.Model

	dtval := time.Date(2000, 1, 2, 3, 4, 5, 6, time.Local)
	xif := new(MockExif)
	xif.On("DateTime").Return(dtval, nil)
	xif.On("LatLong").Return(float64(123), float64(456), nil)
	serial := new(tiff.Tag)
	serial.Val = []byte("shit\u0000")
	xif.On("Get", SerialNumber).Return(serial, nil)
	cmake := new(tiff.Tag)
	cmake.Val = []byte("piss\u0000")
	xif.On("Get", Make).Return(cmake, nil)
	cmodel := new(tiff.Tag)
	cmodel.Val = []byte("fuck\u0000")
	xif.On("Get", Model).Return(cmodel, nil)

	rect := new(image.Rectangle)
	rect.Min = image.Point{X: 0, Y: 0}
	rect.Max = image.Point{X: 1600, Y: 1200}

	p.AddMetadata(rect, xif)

	assert.EqualValues(t, dtval, *p.TakenAt)
	assert.EqualValues(t, float64(123), p.Lat)
	assert.EqualValues(t, float64(456), p.Long)
	assert.EqualValues(t, "shit", p.CamSerial)
	assert.EqualValues(t, "piss", p.CamMake)
	assert.EqualValues(t, "fuck", p.CamModel)
	assert.EqualValues(t, rect.Max.X, p.Width)
	assert.EqualValues(t, rect.Max.Y, p.Height)
	assert.EqualValues(t, float64(1.92), p.Megapixels)
}
