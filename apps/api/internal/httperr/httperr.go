package httperr

import (
	"log"
	"net/http"
)

func Write(w http.ResponseWriter, r *http.Request, status int, err error) {
	if status < http.StatusInternalServerError {
		http.Error(w, err.Error(), status)
		return
	}

	log.Printf("[http] %s %s -> %d: %v", r.Method, r.URL.Path, status, err)
	http.Error(w, publicMessage(status), status)
}

func publicMessage(status int) string {
	switch status {
	case http.StatusBadGateway:
		return "upstream service error"
	case http.StatusServiceUnavailable:
		return "service temporarily unavailable"
	default:
		return "internal server error"
	}
}
