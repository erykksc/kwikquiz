package common

import (
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"github.com/erykksc/kwikquiz/templates"
)

// TmplParseWithBase parses the given template file and base template file
func TmplParseWithBase(path string) *template.Template {
	// Path inside the <project-root>/templates/
	embedPath := strings.TrimPrefix(path, "templates/")

	return template.Must(template.ParseFS(templates.FS, embedPath, "base.html"))
}

func ParseTmplWithFuncs(path string) *template.Template {
	// Path inside the <project-root>/templates/
	embedPath := strings.TrimPrefix(path, "templates/")
	// get base name of the path
	baseName := filepath.Base(path)
	viewTmpl := template.Must(template.New(baseName).Funcs(template.FuncMap{
		"formatAsISO": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		// Decrement function used for checking if the current question is the last one
		"decrement": func(i int) int {
			return i - 1
		},
		"add": func(a, b int) int {
			return a + b
		},
	}).ParseFS(templates.FS, embedPath, "base.html"))

	return viewTmpl
}

var IndexTmpl = TmplParseWithBase("templates/index.html")

// Template for joining a session/lobby
var JoinFormTmpl = IndexTmpl.Lookup("join-form")

type JoinFormData struct {
	GamePinError string
}

var NotFoundTmpl = TmplParseWithBase("templates/not-found.html")
