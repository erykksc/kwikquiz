package quiz

import (
	"html/template"

	"github.com/erykksc/kwikquiz/internal/common"
)

func tmplParseWithBase(path string) *template.Template {
	return template.Must(template.ParseFiles(path, common.BaseTmplPath))
}

// Templates used to render the different pages of the quiz
var QuizzesTemplate = tmplParseWithBase("templates/quizzes/quizzes.html")
var QuizPreviewTemplate = tmplParseWithBase("templates/quizzes/quiz-preview.html")
var QuizCreateTemplate = tmplParseWithBase("templates/quizzes/quiz-create.html")
var QuizUpdateTemplate = tmplParseWithBase("templates/quizzes/quiz-update-2.0.html")
