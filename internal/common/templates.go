package common

import "html/template"

const (
	BaseTmplPath = "templates/base.html"
)

// tmplParseWithBase parses the given template file and base template file
func tmplParseWithBase(path string) *template.Template {
	return template.Must(template.ParseFiles(path, BaseTmplPath))
}

var IndexTmpl = tmplParseWithBase("templates/index.html")

// Template for joining a session/lobby
var JoinFormTmpl = IndexTmpl.Lookup("join-form")

type JoinFormData struct {
	GamePinError string
}

var NotFoundTmpl = tmplParseWithBase("templates/not-found.html")
