package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/kochman/hotshots/config"
)

func TestAuth(t *testing.T) {
	s := &Server{
		cfg: config.Config{
			AuthUsername: "testuname",
			AuthPassword: "testpass",
		},
	}
	r := chi.NewRouter()
	r.Use(s.auth)

	ts := httptest.NewServer(r)
	defer ts.Close()

	type testCase struct {
		username   string
		password   string
		statusCode int
		body       string
	}
	cases := []testCase{
		{
			username:   "testuname",
			password:   "testpass",
			statusCode: 200,
			body:       "hello there",
		},
		{
			username:   "testuname",
			password:   "WRONG",
			statusCode: 401,
			body:       "Unauthorized.\n",
		},
		{
			username:   "WRONG",
			password:   "testpass",
			statusCode: 401,
			body:       "Unauthorized.\n",
		},
		{
			username:   "WRONG",
			password:   "WRONG",
			statusCode: 401,
			body:       "Unauthorized.\n",
		},
	}

	for _, c := range cases {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello there"))
		})

		req, err := http.NewRequest("GET", ts.URL, nil)
		if err != nil {
			t.Errorf("unable to create HTTP request: %s", err)
			continue
		}

		req.SetBasicAuth(c.username, c.password)

		client := http.Client{}
		res, err := client.Do(req)
		if err != nil {
			t.Errorf("unable to get URL: %s", err)
			continue
		}

		bodyBytes, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Errorf("unable to read: %s", err)
			continue
		}
		body := string(bodyBytes)
		if body != c.body {
			t.Errorf("expected body [%s], got [%s]", c.body, body)
		}

		if c.statusCode != res.StatusCode {
			t.Errorf("expected status code %d, got %d", c.statusCode, res.StatusCode)
		}
	}
}

func TestNoAuth(t *testing.T) {
	s := &Server{
		cfg: config.Config{
			AuthUsername: "",
			AuthPassword: "",
		},
	}
	r := chi.NewRouter()
	r.Use(s.auth)

	ts := httptest.NewServer(r)
	defer ts.Close()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello there"))
	})

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Errorf("unable to create HTTP request: %s", err)
		return
	}

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("unable to get URL: %s", err)
		return
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Errorf("unable to read: %s", err)
		return
	}
	body := string(bodyBytes)
	if body != "hello there" {
		t.Errorf("expected body [hello there], got [%s]", body)
	}

	if res.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", res.StatusCode)
	}
}
