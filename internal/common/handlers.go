package common

import (
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
)

const (
	NotFoundPage  = "static/notfound.html"
	TemplateBase  = "templates/base.html"
	TemplateIndex = "templates/index.html"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	tmpl := template.Must(template.ParseFiles(TemplateIndex, TemplateBase))
	tmpl.Execute(w, nil)
}

// ErrorHandler handles HTTP errors based on the status code.
func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {
	switch status {
	case http.StatusNotFound:
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		if len(notFoundPageContent) == 0 {
			loadNotFoundPageContent()
		}
		w.Write(notFoundPageContent)
		return
	default:
		w.WriteHeader(status)
		return
	}
}

var notFoundPageContent []byte

func loadNotFoundPageContent() {
	notFoundFile, err := os.Open(NotFoundPage)
	if err != nil {
		panic(err)
	}
	defer notFoundFile.Close()
	notFoundPageContent, err = io.ReadAll(notFoundFile)
	if err != nil {
		panic(err)
	}
}
