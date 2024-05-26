package quiz

import (
	"fmt"
	"github.com/erykksc/kwikquiz/internal/common"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
)

const (
	NotFoundPage       = "static/notfound.html"
	BaseTemplate       = "templates/base.html"
	IndexTemplate      = "templates/index.html"
	QuizzesTemplate    = "templates/quizzes/quizzes.html"
	QuizTemplate       = "templates/quizzes/quiz-qid.html"
	QuizCreateTemplate = "templates/quizzes/quiz-create-v2.html"
)

var quizzesRepo QuizRepository = NewInMemoryQuizRepository()
var DEBUG = common.DebugOn()

func NewQuizzesRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /quizzes/{$}", getAllQuizzesHandler)
	mux.HandleFunc("GET /quizzes/{qid}", getQuizHandler)
	mux.HandleFunc("POST /quizzes/create/{$}", postQuizHandler)
	mux.HandleFunc("GET /quizzes/create/", getQuizCreateHandler)
	return mux

}

func getAllQuizzesHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	quizzes, err := quizzesRepo.GetAllQuizzes()
	if err != nil {
		common.ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles(QuizzesTemplate, BaseTemplate))

	tmpl.Execute(w, quizzes)
}

func getQuizHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	qidStr := r.PathValue("qid")

	// Convert the string to an integer
	qid, err := strconv.Atoi(qidStr)
	if err != nil {
		// Handle the error if conversion fails
		http.Error(w, "Invalid quid value", http.StatusBadRequest)
		return
	}

	quiz, err := quizzesRepo.GetQuiz(qid)
	if err != nil {
		switch err.(type) {
		case ErrQuizNotFound:
			common.ErrorHandler(w, r, http.StatusNotFound)
			return

		default:
			common.ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}
	tmpl := template.Must(template.ParseFiles(QuizTemplate, BaseTemplate))
	tmpl.Execute(w, quiz)
}

type createQuizForm struct {
	Qid             int
	Title           string
	Description     string
	TimePerQuestion int
	QuestionOrder   string
	Questions       []Question
	FormError       string
}

func postQuizHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	qidStr := r.FormValue("qid")
	title := r.FormValue("title")
	description := r.FormValue("description")
	timePerQuestionStr := r.FormValue("time-per-question")
	questionOrder := r.FormValue("question-order")

	// Convert the string to an integer
	timePerQuestion, err := strconv.Atoi(timePerQuestionStr)
	if err != nil {
		// Handle the error if conversion fails
		http.Error(w, "Invalid time-per-question value", http.StatusBadRequest)
		return
	}
	// Convert the string to an integer
	qid, err := strconv.Atoi(qidStr)
	if err != nil {
		// Handle the error if conversion fails
		http.Error(w, "Invalid quid value", http.StatusBadRequest)
		return
	}
	var questions []Question
	questionIndex := 1
	for {
		questionText := r.FormValue("question-" + strconv.Itoa(questionIndex))
		if questionText == "" {
			break
		}
		answer := []string{
			r.FormValue("answer-" + strconv.Itoa(questionIndex) + "-1"),
			r.FormValue("answer-" + strconv.Itoa(questionIndex) + "-2"),
			r.FormValue("answer-" + strconv.Itoa(questionIndex) + "-3"),
			r.FormValue("answer-" + strconv.Itoa(questionIndex) + "-4"),
		}

		correctAnswerStr := r.FormValue("correct-answer-" + strconv.Itoa(questionIndex))
		correctAnswer, err := strconv.Atoi(correctAnswerStr)
		if err != nil {
			http.Error(w, "Invalid answer option value", http.StatusBadRequest)
			return
		}

		questions = append(questions, Question{
			Text:          questionText,
			Answers:       answer,
			CorrectAnswer: correctAnswer,
		})
		fmt.Println("%+v", questions)
		questionIndex++
	}
	// Create new quiz
	quiz := Quiz{
		ID:              qid,
		Title:           title,
		Description:     description,
		TimePerQuestion: timePerQuestion,
		QuestionOrder:   questionOrder,
		Questions:       questions,
	}
	fmt.Println("%+v", quiz)

	if err := quizzesRepo.AddQuiz(quiz); err != nil {
		slog.Error("Error adding quiz", "error", err)
		tmpl := template.Must(template.ParseFiles(QuizCreateTemplate, BaseTemplate))
		err := tmpl.ExecuteTemplate(w, "create-form", createQuizForm{
			Qid:             qid,
			Title:           title,
			Description:     description,
			TimePerQuestion: timePerQuestion,
			QuestionOrder:   questionOrder,
			Questions:       questions,
		})
		if err != nil {
			slog.Error("Error rendering quiz", "error", err)
		}
		return
	}

	// Redirecting to the quiz
	w.Header().Add("HX-Redirect", "/quizzes/"+qidStr)
	w.WriteHeader(http.StatusCreated)

}

func getQuizCreateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	tmpl := template.Must(template.ParseFiles(QuizCreateTemplate, BaseTemplate))
	tmpl.Execute(w, nil)
}
