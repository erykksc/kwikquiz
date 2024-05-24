package quiz

import (
	"github.com/erykksc/kwikquiz/internal/common"
	"html/template"
	"log/slog"
	"net/http"
)

const (
	NotFoundPage       = "static/notfound.html"
	BaseTemplate       = "templates/base.html"
	IndexTemplate      = "templates/index.html"
	QuizzesTemplate    = "templates/quizzes/quizzes.html"
	QuizTemplate       = ""
	QuizCreateTemplate = "templates/quizzes/quiz-create.html"
)

var quizzesRepo QuizRepository = NewInMemoryQuizRepository()
var DEBUG = common.DebugOn()

func NewQuizzesRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /quizzes/{$}", getAllQuizzesHandler)
	mux.HandleFunc("GET /quizzes/{qid}", getQuizHandler)
	mux.HandleFunc("POST /quizzes/{$}", postQuizHandler)
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
	qid := r.PostFormValue("gid")

	quiz, err := quizzesRepo.GetQuiz(QuizID(qid))
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
	Qid         string
	Title       string
	Description string
	Owner       string
	FormError   string
}

func postQuizHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	qid := r.FormValue("qid")
	title := r.FormValue("title")
	description := r.FormValue("description")
	owner := r.FormValue("owner")

	if qid == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("qid in form is required"))
		return
	}

	if title == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("title in form is required"))
		return
	}

	if description == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("description in form is required"))
		return
	}
	if owner == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("owner in form is required"))
		return
	}

	// Create new quiz
	quiz := Quiz{
		ID:          QuizID(qid),
		Title:       title,
		Description: description,
		Owner:       owner,
	}

	if err := quizzesRepo.AddQuiz(quiz); err != nil {
		slog.Error("Error adding quiz", "error", err)
		tmpl := template.Must(template.ParseFiles(QuizCreateTemplate, BaseTemplate))
		err := tmpl.ExecuteTemplate(w, "create-form", createQuizForm{
			Qid:         qid,
			Title:       title,
			Description: description,
			Owner:       owner,
		})
		if err != nil {
			slog.Error("Error rendering quiz", "error", err)
		}
		return
	}

	// Redirecting to the quiz
	w.Header().Add("HX-Redirect", "/quizzes/"+qid)
	w.WriteHeader(http.StatusCreated)
}

func getQuizCreateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	tmpl := template.Must(template.ParseFiles(QuizCreateTemplate, BaseTemplate))
	tmpl.Execute(w, nil)
}
