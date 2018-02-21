package server

import (
	"encoding/json"
	"math"
	"net/http"
)

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func WriteError(s string, status int, w http.ResponseWriter) {
	w.WriteHeader(status)
	v := ErrorResponse{
		Success: false,
		Error:   s,
	}
	WriteJsonResponse(v, w)
}

func WriteJsonResponse(v interface{}, w http.ResponseWriter) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
