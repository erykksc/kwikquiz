package quiz

import "github.com/erykksc/kwikquiz/internal/common"

// Templates used to render the different pages of the quiz
var QuizzesTemplate = common.TmplParseWithBase("templates/quizzes/quizzes.html")
var QuizPreviewTemplate = common.TmplParseWithBase("templates/quizzes/quiz-preview.html")
var QuizCreateTemplate = common.TmplParseWithBase("templates/quizzes/quiz-create.html")
var QuizUpdateTemplate = common.TmplParseWithBase("templates/quizzes/quiz-update-2.0.html")
