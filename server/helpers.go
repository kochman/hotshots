package server

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/index"
)

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func GetPaginateValues(r *http.Request) (func(*index.Options), func(*index.Options), error) {
	var start, limit func(*index.Options)
	startVal := r.URL.Query().Get("start")
	limitVal := r.URL.Query().Get("limit")

	if startVal != "" {
		startVal, err := strconv.ParseInt(startVal, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		start = storm.Skip(int(startVal))
	} else {
		start = func(*index.Options) {}
	}

	if limitVal != "" {
		limitVal, err := strconv.ParseInt(limitVal, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		limit = storm.Limit(int(limitVal))
	} else {
		limit = storm.Limit(20)
	}

	return start, limit, nil
}

func WriteError(s string, status int, w http.ResponseWriter) {
	v := ErrorResponse{
		Success: false,
		Error:   s,
	}
	WriteJsonResponse(v, status, w)
}

func WriteJsonResponse(v interface{}, status int, w http.ResponseWriter) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()
	select {
	case <-c:
		return false // compl eted normally
	case <-time.After(timeout):
		return true // timed out
	}
}
