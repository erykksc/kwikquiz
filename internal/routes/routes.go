package routes

import (
	"fmt"
	"html/template"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html", "templates/base.html"))
		tmpl.ExecuteTemplate(w, "base", nil)
	})
	mux.HandleFunc("POST /session", postSessionHandler)
	// Serve static files from the public directory
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	return mux
}

func postSessionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "POST /session")
}
