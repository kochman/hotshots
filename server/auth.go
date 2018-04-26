package server

import (
	"crypto/subtle"
	"net/http"
)

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(s.cfg.AuthUsername) == 0 && len(s.cfg.AuthPassword) == 0 {
			// Authentication is not configured
			next.ServeHTTP(w, r)
			return
		}

		username, password, _ := r.BasicAuth()

		okUsername := subtle.ConstantTimeCompare([]byte(s.cfg.AuthUsername), []byte(username)) == 1
		okPassword := subtle.ConstantTimeCompare([]byte(s.cfg.AuthPassword), []byte(password)) == 1

		if okUsername && okPassword {
			next.ServeHTTP(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", `Basic realm="Hotshots"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		}
	})
}
