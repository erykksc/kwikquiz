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
var QuizCreateTemplate *template.Template
var QuizUpdateTemplate *template.Template

func init() {
	QuizzesTemplate = tmplParseWithBase("templates/quizzes/quizzes.html")
	QuizPreviewTemplate = tmplParseWithBase("templates/quiz/quiz-preview.html")
	QuizCreateTemplate = tmplParseWithBase("templates/quiz/quiz-create.html")
	QuizUpdateTemplate = tmplParseWithBase("templates/quiz/quiz-update.html")
}
