package common

import (
	"html/template"
	"path/filepath"
	"time"

	embed_files "github.com/erykksc/kwikquiz"
)

const (
	BaseTmplPath = "templates/base.html"
)

// TmplParseWithBase parses the given template file and base template file
func TmplParseWithBase(path string) *template.Template {
	return template.Must(template.ParseFS(embed_files.Templates, path, BaseTmplPath))
}

func ParseTmplWithFuncs(path string) *template.Template {
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
	}).ParseFS(embed_files.Templates, path, BaseTmplPath))

	return viewTmpl
}

var IndexTmpl = TmplParseWithBase("templates/index.html")

// Template for joining a session/lobby
var JoinFormTmpl = IndexTmpl.Lookup("join-form")

type JoinFormData struct {
	GamePinError string
}

var NotFoundTmpl = TmplParseWithBase("templates/not-found.html")
