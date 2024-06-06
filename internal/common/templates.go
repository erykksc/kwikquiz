package common

import "html/template"

const (
	BaseTmplPath = "templates/base.html"
)

var IndexTmpl *template.Template
var NotFoundTmpl *template.Template

func init() {
	IndexTmpl = template.Must(template.ParseFiles("templates/index.html", BaseTmplPath))
	NotFoundTmpl = template.Must(template.ParseFiles("templates/not-found.html", BaseTmplPath))
}
