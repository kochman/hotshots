package pusher

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
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
		t.Errorf("got %d existing photos, expected 2", len(existing))
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
