package quiz

import (
	"html/template"
	"path/filepath"
	"strings"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/templates"
)

var funcMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
}

var QuizzesTemplate = common.TmplParseWithBase("templates/quizzes/quizzes.html")
var QuizPreviewTemplate = common.TmplParseWithBase("templates/quizzes/quiz-preview.html")

func parseWithFuncs(path string) *template.Template {
	embedPath := strings.TrimPrefix(path, "templates/")
	baseName := filepath.Base(embedPath)
	return template.Must(
		template.New(baseName).Funcs(funcMap).ParseFS(
			templates.FS, embedPath, "quizzes/question-list-partial.html", "base.html",
		),
	)
}

var QuizCreateTemplate = parseWithFuncs("templates/quizzes/quiz-create.html")
var QuizUpdateTemplate = parseWithFuncs("templates/quizzes/quiz-update-2.0.html")

var QuestionListTmpl = template.Must(
	template.New("question-list-partial.html").Funcs(funcMap).ParseFS(
		templates.FS, "quizzes/question-list-partial.html",
	),
)

type QuestionListData struct {
	Questions    []Question
	ActionPrefix string
}
