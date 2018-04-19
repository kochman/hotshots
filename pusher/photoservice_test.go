package pusher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/kochman/hotshots/server"
)

func TestExistingPhotosNone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
{
    "success": true,
    "ids": []
}
    `)
	}))
	defer server.Close()

	ps := &remoteAPI{
		url: server.URL,
	}

	existing, err := ps.existingPhotos()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if len(existing) != 0 {
		t.Errorf("got %d existing photos, expected 0", len(existing))
	}
}

func TestExistingPhotos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `
{
    "success": true,
    "ids": [
        "6e71b3cac15d32fe2d36c270887df9479c25c640",
        "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
        "490528f36debf7c15cea5e9a9d1ea024cf6b2921"
    ]
}
    `)
	}))
	defer server.Close()

	ps := &remoteAPI{
		url: server.URL,
	}

	existing, err := ps.existingPhotos()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if len(existing) != 3 {
		t.Errorf("got %d existing photos, expected 3", len(existing))
	}

	if existing[0] != "6e71b3cac15d32fe2d36c270887df9479c25c640" {
		t.Errorf("got unexpected photo ID")
	}
	if existing[1] != "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d" {
		t.Errorf("got unexpected photo ID")
	}
	if existing[2] != "490528f36debf7c15cea5e9a9d1ea024cf6b2921" {
		t.Errorf("got unexpected photo ID")
	}
}

func TestExistingPhotosPagination(t *testing.T) {
	type testPhotoID struct {
		id         string
		numReturns int
	}

	testPagination := func(numIDs int) {
		ids := []testPhotoID{}
		for i := 0; i < numIDs; i++ {
			testID := testPhotoID{
				id:         string(i),
				numReturns: 0,
			}
			ids = append(ids, testID)
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("start") == "" {
				t.Errorf("missing start parameter")
				return
			}
			start, err := strconv.Atoi(r.URL.Query().Get("start"))
			if err != nil {
				t.Errorf("unable to convert start to integer: %s", err)
				return
			}

			if r.URL.Query().Get("limit") == "" {
				t.Errorf("missing limit parameter")
				return
			}
			limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
			if err != nil {
				t.Errorf("unable to convert start to integer: %s", err)
				return
			}

			resp := server.GetPhotoIDsResponse{
				Success: true,
				IDs:     []string{},
			}
			for i := start; i < start+limit; i++ {
				if i >= len(ids) {
					break
				}
				resp.IDs = append(resp.IDs, ids[i].id)
				ids[i].numReturns++
			}

			body, err := json.Marshal(resp)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(body)
		}))
		defer server.Close()

		ps := &remoteAPI{
			url: server.URL,
		}

		existing, err := ps.existingPhotos()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}
		if len(existing) != len(ids) {
			t.Errorf("got %d existing photos, expected %d", len(existing), len(ids))
		}

		for i, testID := range ids {
			if testID.id != existing[i] {
				t.Errorf("IDs do not match")
			}
			if testID.numReturns != 1 {
				t.Errorf("got %d returns, expected 1", testID.numReturns)
			}
		}
	}

	testPagination(0)
	testPagination(1)
	testPagination(99)
	testPagination(100)
	testPagination(101)
}

func TestUploadPhoto(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("got unexpected request method %s", r.Method)
			return
		}
		if r.Header.Get("Content-Type")[:21] != "multipart/form-data; " {
			t.Errorf("got unexpected Content-Type: %s", r.Header.Get("Content-Type"))
			return
		}

		err := r.ParseMultipartForm(32 << 20) // 32 MB
		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		field, ok := r.MultipartForm.File["photo"]
		if !ok {
			t.Errorf("photo field not in request form")
			return
		}

		if len(field) != 1 {
			t.Errorf("got %d fields, expected 1", len(field))
			return
		}

		file, err := field[0].Open()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		contents, err := ioutil.ReadAll(file)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		if string(contents) != "hello" {
			t.Errorf("got unexpected photo")
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"success": true}`)
	}))
	defer server.Close()

	ps := &remoteAPI{
		url: server.URL,
	}

	err := ps.uploadPhoto([]byte("hello"))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
}

func TestUploadPhotoExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"success": false, "error": "photo already exists"}`)
	}))
	defer server.Close()

	ps := &remoteAPI{
		url: server.URL,
	}

	err := ps.uploadPhoto([]byte("hello"))
	if err != errPhotoExists {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestUploadPhotoUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	ps := &remoteAPI{
		url: server.URL,
	}

	err := ps.uploadPhoto([]byte("hello"))
	if err.Error() != "unexpected status code: 500" {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestUploadPhotoUnsuccessfulResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"success": false}`)
	}))
	defer server.Close()

	ps := &remoteAPI{
		url: server.URL,
	}

	err := ps.uploadPhoto([]byte("hello"))
	if err.Error() != "unsuccessful response" {
		t.Errorf("unexpected error: %s", err)
	}
}
