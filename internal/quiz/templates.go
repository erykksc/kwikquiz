package quiz

import (
	"html/template"

	"github.com/erykksc/kwikquiz/internal/common"
)

func tmplParseWithBase(path string) *template.Template {
	return template.Must(template.ParseFiles(path, common.BaseTmplPath))
}

// Templates used to render the different pages of the quiz
var QuizzesTemplate *template.Template
var QuizPreviewTemplate *template.Template
var QuizCreateUpdateTemplate *template.Template

func init() {
	QuizzesTemplate = tmplParseWithBase("templates/quizzes/quizzes.html")
	QuizPreviewTemplate = tmplParseWithBase("templates/quizzes/quiz-preview.html")
	QuizCreateUpdateTemplate = tmplParseWithBase("templates/quizzes/quiz-create-update.html")
}
