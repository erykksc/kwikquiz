package quiz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/database"
	"github.com/erykksc/kwikquiz/internal/models"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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
	//if common.DevMode() {
	//	QuizzesRepo.AddQuiz(ExampleQuizGeography)
	//	slog.Info("Added example geography quiz")
	//	QuizzesRepo.AddQuiz(ExampleQuizMath)
	//	slog.Info("Added example math quiz")
	//}

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
	var qid uint

	// Convert the string to an integer

	// Convert the string to an integer if qidStr is not empty
	if qidStr != "" {
		qidInt, convErr := strconv.Atoi(qidStr)
		if convErr != nil {
			slog.Error("Error converting qid", "error", convErr)
			return models.Quiz{}, fmt.Errorf("invalid quiz ID")
		}
		qid = uint(qidInt)
	}
	title := r.FormValue("title")
	password := r.FormValue("password")
	description := r.FormValue("description")
	questions, err := parseQuestions(r)
	if err != nil {
		return models.Quiz{}, err
	}

	// Parse questions
	questions, parseErr := parseQuestions(r)
	if parseErr != nil {
		return models.Quiz{}, parseErr
	}

	// Return the Quiz model based on whether qid is provided or not
	return models.Quiz{
		ID:          qid,
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

		var answers []models.Answer
		answerIndex := 1
		for {
			answerPrefix := "answer-" + strconv.Itoa(questionIndex) + "-" + strconv.Itoa(answerIndex)

			// Check if any input with this name exists
			_, _, err := r.FormFile(answerPrefix)
			textValue := r.FormValue(answerPrefix)

			if err != nil && textValue == "" {
				break // No more answers for this question
			}

			var answer models.Answer

			// Determine answer type based on the input field present
			if textValue != "" {
				// Check if it's a textarea (LaTeX) or text input
				if strings.Contains(textValue, "\n") {
					answer = models.Answer{
						LaTeX: textValue,
					}
				} else {
					answer = models.Answer{
						Text: textValue,
					}
				}
			} else {
				// It's an image file
				file, header, err := r.FormFile(answerPrefix)
				if err != nil {
					return nil, fmt.Errorf("error reading image file: %v", err)
				}
				defer file.Close()

				var buf bytes.Buffer
				_, err = io.Copy(&buf, file)
				if err != nil {
					return nil, fmt.Errorf("error copying image file: %v", err)
				}

				answer = models.Answer{
					Image:     buf.Bytes(),
					ImageName: header.Filename,
				}
			}

			// Check if the answer is correct
			correctBtnValue := r.FormValue("correct-answer-" + strconv.Itoa(questionIndex) + "-" + strconv.Itoa(answerIndex))
			answer.IsCorrect = correctBtnValue == "Correct"

			answers = append(answers, answer)
			answerIndex++
		}

		questions = append(questions, models.Question{
			Text:    questionText,
			Answers: answers,
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
