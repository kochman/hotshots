package server

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	PageSize = 20
)

type TagMatcher struct {
	tag string
}

func (m *TagMatcher) MatchField(v interface{}) (bool, error) {
	tags, ok := v.([]string)
	if !ok {
		return false, errors.New("failed to convert field")
	}
	for _, t := range tags {
		if strings.Contains(t, m.tag) {
			return true, nil
		}
	}
	return false, nil
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func removeElement(list []string, loc int) []string {
	// Does not return in order
	list[loc] = list[len(list)-1]
	return list[:len(list)-1]
}

func GetPaginateValues(r *http.Request) (int, int, error) {
	var start, limit int
	startStr := r.URL.Query().Get("start")
	limitStr := r.URL.Query().Get("limit")

	if startStr != "" {
		startVal, err := strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		start = int(startVal)
	} else {
		start = 0
	}

	if limitStr != "" {
		limitVal, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		limit = int(limitVal)
	} else {
		limit = PageSize
	}

	return start, limit, nil
}

func GetDeleted(r *http.Request) (bool, error) {
	var deleted bool
	deletedStr := r.URL.Query().Get("deleted")

	if deletedStr != "" {
		deletedVar, err := strconv.ParseBool(deletedStr)
		if err != nil {
			return false, err
		}
		deleted = deletedVar
	} else {
		deleted = false
	}

	return deleted, nil
}

func GetTag(r *http.Request) string {
	return r.URL.Query().Get("tag")
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
