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
	QuizCreateTemplate = "templates/quizzes/quiz-create.html"
)

var quizzesRepo QuizRepository = NewInMemoryQuizRepository()
var DEBUG = common.DebugOn()

func NewQuizzesRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /quizzes/{$}", getAllQuizzesHandler)
	mux.HandleFunc("GET /quizzes/{qid}", getQuizHandler)
	mux.HandleFunc("POST /quizzes/create/{$}", postQuizHandler)
	mux.HandleFunc("GET /quizzes/create/", getQuizCreateHandler)
	mux.HandleFunc("GET /quizzes/create", getQuizCreateHandler)
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
		http.Error(w, "Invalid qid value", http.StatusBadRequest)
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
	fmt.Println(quiz)
}

type createQuizForm struct {
	Qid           int
	Title         string
	Description   string
	QuestionOrder string
	Questions     []Question
	FormError     string
}

func postQuizHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	quiz, err := parseQuizForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	quizID, err := quizzesRepo.AddQuiz(quiz)
	if err != nil {
		slog.Error("Error adding quiz", "error", err)
		renderQuizCreateForm(w, quiz, err)
		return
	}

	redirectToQuiz(w, quizID)
}

func parseQuizForm(r *http.Request) (Quiz, error) {
	title := r.FormValue("title")
	password := r.FormValue("password")
	description := r.FormValue("description")
	questionOrder := r.FormValue("question-order")

	questions, err := parseQuestions(r)
	if err != nil {
		return Quiz{}, err
	}

	return Quiz{
		Title:         title,
		Password:      password,
		Description:   description,
		QuestionOrder: questionOrder,
		Questions:     questions,
	}, nil
}

func parseQuestions(r *http.Request) ([]Question, error) {
	var questions []Question
	questionIndex := 1

	for {
		questionText := r.FormValue("question-" + strconv.Itoa(questionIndex))
		if questionText == "" {
			break
		}

		answers := []string{
			r.FormValue("answer-" + strconv.Itoa(questionIndex) + "-1"),
			r.FormValue("answer-" + strconv.Itoa(questionIndex) + "-2"),
			r.FormValue("answer-" + strconv.Itoa(questionIndex) + "-3"),
			r.FormValue("answer-" + strconv.Itoa(questionIndex) + "-4"),
		}

		correctAnswerStr := r.FormValue("correct-answer-" + strconv.Itoa(questionIndex))
		correctAnswer, err := strconv.Atoi(correctAnswerStr)
		if err != nil {
			return nil, fmt.Errorf("invalid answer option value")
		}

		questions = append(questions, Question{
			Text:          questionText,
			Answers:       answers,
			CorrectAnswer: correctAnswer,
		})
		questionIndex++
	}
	return questions, nil
}

func renderQuizCreateForm(w http.ResponseWriter, quiz Quiz, err error) {
	tmpl := template.Must(template.ParseFiles(QuizCreateTemplate, BaseTemplate))
	tmpl.ExecuteTemplate(w, "create-form", createQuizForm{
		Title:         quiz.Title,
		Description:   quiz.Description,
		QuestionOrder: quiz.QuestionOrder,
		Questions:     quiz.Questions,
		FormError:     err.Error(),
	})
}

func redirectToQuiz(w http.ResponseWriter, quizID int) {
	w.Header().Add("HX-Redirect", fmt.Sprintf("/quizzes/%d", quizID))
	w.WriteHeader(http.StatusCreated)
}

func getQuizCreateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	tmpl := template.Must(template.ParseFiles(QuizCreateTemplate, BaseTemplate))
	tmpl.Execute(w, nil)
}
