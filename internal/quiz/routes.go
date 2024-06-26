package quiz

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/database"
	"github.com/erykksc/kwikquiz/internal/models"
	"log/slog"
	"net/http"
	"strconv"
)

var QuizzesRepo *GormQuizRepository

func NewQuizzesRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /quizzes/{$}", getAllQuizzesHandler)
	mux.HandleFunc("GET /quizzes/{qid}", getQuizHandler)
	mux.HandleFunc("POST /quizzes/create/{$}", postQuizHandler)
	mux.HandleFunc("GET /quizzes/create/{$}", getQuizCreateHandler)
	mux.HandleFunc("GET /quizzes/update/{qid}", getQuizUpdateHandler)
	mux.HandleFunc("PUT /quizzes/update/{qid}", updateQuizHandler)
	mux.HandleFunc("DELETE /quizzes/delete/{qid}", deleteQuizHandler)

	// init database instance
	QuizzesRepo = NewGormQuizRepository(database.DB)

	// Add quiz if in debug mode
	if common.DevMode() {
		QuizzesRepo.AddQuiz(ExampleQuizGeography)
		slog.Info("Added example geography quiz")
		QuizzesRepo.AddQuiz(ExampleQuizMath)
		slog.Info("Added example math quiz")
	}

	return mux
}

func getAllQuizzesHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	quizzes, err := QuizzesRepo.GetAllQuizzes()
	if err != nil {
		common.ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	err = QuizzesTemplate.Execute(w, quizzes)
	if err != nil {
		slog.Error("Error rendering template", "err", err)
	}
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

	quiz, err := QuizzesRepo.GetQuiz(uint(qid))
	if err != nil {
		var errQuizNotFound ErrQuizNotFound
		switch {
		case errors.As(err, &errQuizNotFound):
			common.ErrorHandler(w, r, http.StatusNotFound)
			return
		default:
			common.ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}
	err = QuizPreviewTemplate.Execute(w, quiz)
	if err != nil {
		slog.Error("Error getting quiz..", "err", err)
	}
}

type createQuizForm struct {
	Qid         uint
	Title       string
	Password    string
	Description string
	Questions   []models.Question
	FormError   string
}

func postQuizHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	quiz, err := parseQuizForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	quizID, err := QuizzesRepo.AddQuiz(quiz)
	if err != nil {
		slog.Error("Error adding quiz", "error", err)
		renderQuizCreateForm(w, quiz, err)
		return
	}

	redirectToQuiz(w, quizID)
}

func parseQuizForm(r *http.Request) (models.Quiz, error) {
	qidStr := r.PathValue("qid")

	// Convert the string to an integer
	qid, err := strconv.Atoi(qidStr)
	if err != nil {
		// Handle the error if conversion fails
		slog.Error("Error converting qid", "error", err)
	}
	title := r.FormValue("title")
	password := r.FormValue("password")
	description := r.FormValue("description")
	questions, err := parseQuestions(r)
	if err != nil {
		return models.Quiz{}, err
	}

	return models.Quiz{
		ID:          uint(qid),
		Title:       title,
		Password:    password,
		Description: description,
		Questions:   questions,
	}, nil
}

func parseQuestions(r *http.Request) ([]models.Question, error) {
	var questions []models.Question
	questionIndex := 1

	for {
		questionText := r.FormValue("question-" + strconv.Itoa(questionIndex))
		if questionText == "" {
			break
		}
		// Get the correct answer string
		correctAnswerStr := r.FormValue("correct-answer-" + strconv.Itoa(questionIndex))
		correctAnswer, err := strconv.Atoi(correctAnswerStr)
		if err != nil {
			return nil, fmt.Errorf("invalid answer option value")
		}
		// Append answers to a slice
		var answers []models.Answer
		for answerIndex := 1; answerIndex <= 4; answerIndex++ {
			answerText := r.FormValue("answer-" + strconv.Itoa(questionIndex) + "-" + strconv.Itoa(answerIndex))
			if answerText == "" {
				return nil, fmt.Errorf("missing answer text for question %d, answer %d", questionIndex, answerIndex)
			}
			answers = append(answers, models.Answer{
				IsCorrect: answerIndex == correctAnswer,
				Text:      answerText,
			})

		}
		// Append questions to a slice
		questions = append(questions, models.Question{
			Text:          questionText,
			Answers:       answers,
			CorrectAnswer: correctAnswer,
		})
		questionIndex++
	}
	return questions, nil
}

func renderQuizCreateForm(w http.ResponseWriter, quiz models.Quiz, err error) {
	err = QuizCreateTemplate.ExecuteTemplate(w, "create-form", createQuizForm{
		Title:       quiz.Title,
		Description: quiz.Description,
		Questions:   quiz.Questions,
		FormError:   err.Error(),
	})
	if err != nil {
		slog.Error("Error rendering template", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func redirectToQuiz(w http.ResponseWriter, quizID uint) {
	w.Header().Add("HX-Redirect", fmt.Sprintf("/quizzes/%d", quizID))
	w.WriteHeader(http.StatusCreated)
}

func getQuizCreateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	err := QuizCreateTemplate.Execute(w, nil)
	if err != nil {
		slog.Error("Error rendering template", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getQuizUpdateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	qidStr := r.PathValue("qid")

	// Convert the string to an integer
	qid, err := strconv.Atoi(qidStr)
	if err != nil {
		// Handle the error if conversion fails
		http.Error(w, "Invalid qid value", http.StatusBadRequest)
		return
	}

	quiz, err := QuizzesRepo.GetQuiz(uint(qid))
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

	// Serialize quiz data to JSON
	quizJSON, err := json.Marshal(quiz)
	if err != nil {
		slog.Error("Error marshaling quiz data to JSON", "err", err)
		http.Error(w, "Error processing quiz data", http.StatusInternalServerError)
		return
	}

	err = QuizUpdateTemplate.Execute(w, map[string]interface{}{
		"Quiz":     quiz,
		"QuizJSON": string(quizJSON),
	})
	if err != nil {
		slog.Error("Error rendering template", "err", err)
		http.Error(w, "Error executing templates", http.StatusInternalServerError)
		return
	}
}

func updateQuizHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	quiz, err := parseQuizForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	quizID, err := QuizzesRepo.UpdateQuiz(quiz)
	if err != nil {
		slog.Error("Error adding quiz", "error", err)
		renderQuizCreateForm(w, quiz, err)
		return
	}

	redirectToQuiz(w, quizID)
}

func deleteQuizHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	qidStr := r.PathValue("qid")

	// Convert the string to an integer
	qid, err := strconv.Atoi(qidStr)
	if err != nil {
		// Handle the error if conversion fails
		http.Error(w, "Invalid qid value", http.StatusBadRequest)
		return
	}

	err = QuizzesRepo.DeleteQuiz(uint(qid))
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
	slog.Info("Quiz deleted", "qid", qid)
	w.Header().Add("HX-Redirect", fmt.Sprintf("/"))
	w.WriteHeader(http.StatusNoContent)
}
