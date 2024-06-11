package common

import "html/template"

const (
	BaseTmplPath = "templates/base.html"
)

var IndexTmpl *template.Template

// Template for joining a session/lobby
var JoinFormTmpl *template.Template
var NotFoundTmpl *template.Template

func init() {
	IndexTmpl = template.Must(template.ParseFiles("templates/index.html", BaseTmplPath))
	JoinFormTmpl = IndexTmpl.Lookup("join-form")
	NotFoundTmpl = template.Must(template.ParseFiles("templates/not-found.html", BaseTmplPath))
}
