package server

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const emptyPhotoJSON = `{"id":"","uploaded_at":null,"taken_at":null,"width":0,"height":0,"megapixels":0,"lat":0,
"long":0,"cam_serial":"","cam_make":"","cam_model":"","status":"processing","status_updated_at":null}`

const emptyJSON = `{}`

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
