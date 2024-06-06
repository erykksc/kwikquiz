package common

import (
	"log/slog"
	"net/http"
)

// ErrorHandler handles HTTP errors based on the status code.
func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {
	switch status {
	case http.StatusNotFound:
		NotFoundHandler(w, r)
		return
	default:
		w.WriteHeader(status)
		return
	}
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	if err := NotFoundTmpl.Execute(w, nil); err != nil {
		slog.Error("Error rendering template", "error", err)
	}
}
