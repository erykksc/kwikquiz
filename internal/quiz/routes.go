package quiz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/erykksc/kwikquiz/internal/common"
)

func (s Service) NewQuizzesRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /quizzes/{$}", s.getAllQuizzesHandler)
	mux.HandleFunc("GET /quizzes/{qid}", s.getQuizHandler)
	mux.HandleFunc("POST /quizzes/create/{$}", s.postQuizHandler)
	mux.HandleFunc("GET /quizzes/create/{$}", s.getQuizCreateHandler)
	mux.HandleFunc("GET /quizzes/update/{qid}", s.getQuizUpdateHandler)
	mux.HandleFunc("PUT /quizzes/update/{qid}", s.updateQuizHandler)
	mux.HandleFunc("DELETE /quizzes/delete/{qid}", s.deleteQuizHandler)

	return mux
}

func (s Service) getAllQuizzesHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	quizzes, err := s.repo.GetAll()
	if err != nil {
		common.ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	err = QuizzesTemplate.Execute(w, quizzes)
	if err != nil {
		slog.Error("Error rendering template", "err", err)
	}
}

func (s Service) getQuizHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	qidStr := r.PathValue("qid")

	// Convert the string to an integer
	qid, err := strconv.Atoi(qidStr)
	if err != nil {
		// Handle the error if conversion fails
		http.Error(w, "Invalid qid value", http.StatusBadRequest)
		return
	}

	quiz, err := s.repo.Get(int64(qid))
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
	Questions   []Question
	FormError   string
}

func (s Service) postQuizHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	quiz, err := s.parseQuizForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = s.repo.Insert(&quiz)
	if err != nil {
		slog.Error("Error adding quiz", "error", err)
		s.renderQuizCreateForm(w, quiz, err)
		return
	}
	var lobbyPin = r.FormValue("lobbyPin")

	s.redirectToQuiz(w, lobbyPin)
}

func (s Service) parseQuizForm(r *http.Request) (Quiz, error) {
	qidStr := r.PathValue("qid")
	var qid int64

	// Convert the string to an integer if qidStr is not empty
	if qidStr != "" {
		qidInt, convErr := strconv.Atoi(qidStr)
		if convErr != nil {
			slog.Error("Error converting qid", "error", convErr)
			return Quiz{}, fmt.Errorf("invalid quiz ID")
		}
		qid = int64(qidInt)
	}
	title := r.FormValue("title")
	password := r.FormValue("password")
	description := r.FormValue("description")

	// Parse questions
	questions, err := s.parseQuestions(r)
	if err != nil {
		return Quiz{}, err
	}

	// Return the Quiz model based on whether qid is provided or not
	return Quiz{
		ID:          qid,
		TitleField:  title,
		Password:    password,
		Description: description,
		Questions:   questions,
	}, nil
}

func (s Service) parseQuestions(r *http.Request) ([]Question, error) {
	var questions []Question
	questionIndex := 1

	for {
		questionText := r.FormValue("question-" + strconv.Itoa(questionIndex))
		if questionText == "" {
			break
		}

		var answers []Answer
		answerIndex := 1
		for {
			answerPrefix := "answer-" + strconv.Itoa(questionIndex) + "-" + strconv.Itoa(answerIndex)

			// Check if any input with this name exists
			_, _, err := r.FormFile(answerPrefix)
			textValue := r.FormValue(answerPrefix)

			if err != nil && textValue == "" {
				break // No more answers for this question
			}

			var answer Answer

			// Determine answer type based on the input field present
			if textValue != "" {
				// Check if it's a textarea (LaTeX) or text input
				if strings.Contains(textValue, "\n") {
					answer = Answer{
						LaTeX: textValue,
					}
				} else {
					answer = Answer{
						TextField: textValue,
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

				answer = Answer{
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

		questions = append(questions, Question{
			Text:    questionText,
			answers: answers,
		})
		questionIndex++
	}
	return questions, nil
}

func (s Service) renderQuizCreateForm(w http.ResponseWriter, quiz Quiz, err error) {
	err = QuizCreateTemplate.ExecuteTemplate(w, "create-form", createQuizForm{
		Title:       quiz.TitleField,
		Description: quiz.Description,
		Questions:   quiz.Questions,
		FormError:   err.Error(),
	})
	if err != nil {
		slog.Error("Error rendering template", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s Service) redirectToQuiz(w http.ResponseWriter, lobbyPin string) {
	w.Header().Add("HX-Redirect", fmt.Sprintf("/lobbies/%s", lobbyPin))
	w.WriteHeader(http.StatusCreated)
}

func (s Service) getQuizCreateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	// Extract the 'LobbyPin' query parameter
	queryParams := r.URL.Query()
	lobbyPin := queryParams.Get("LobbyPin")

	// Create data to pass to the template
	data := struct {
		LobbyPin string
	}{
		LobbyPin: lobbyPin,
	}

	err := QuizCreateTemplate.Execute(w, data)
	if err != nil {
		slog.Error("Error rendering template", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s Service) getQuizUpdateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	qidStr := r.PathValue("qid")

	// Convert the string to an integer
	qid, err := strconv.Atoi(qidStr)
	if err != nil {
		// Handle the error if conversion fails
		http.Error(w, "Invalid qid value", http.StatusBadRequest)
		return
	}

	quiz, err := s.repo.Get(int64(qid))
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

	// Extract the 'LobbyPin' query parameter
	queryParams := r.URL.Query()
	lobbyPin := queryParams.Get("LobbyPin")

	// Create data to pass to the template
	data := struct {
		Pin string
	}{
		Pin: lobbyPin,
	}

	err = QuizUpdateTemplate.Execute(w, map[string]interface{}{
		"Quiz":     quiz,
		"QuizJSON": string(quizJSON),
		"LobbyPin": data,
	})
	if err != nil {
		slog.Error("Error rendering template", "err", err)
		http.Error(w, "Error executing templates", http.StatusInternalServerError)
		return
	}
}

func (s Service) updateQuizHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	quiz, err := s.parseQuizForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = s.repo.Update(&quiz)
	if err != nil {
		slog.Error("Error adding quiz", "error", err)
		s.renderQuizCreateForm(w, quiz, err)
		return
	}
	var lobbyPin = r.FormValue("lobbyPin")

	s.redirectToQuiz(w, lobbyPin)
}

func (s Service) deleteQuizHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)

	qidStr := r.PathValue("qid")

	// Convert the string to an integer
	qid, err := strconv.Atoi(qidStr)
	if err != nil {
		// Handle the error if conversion fails
		http.Error(w, "Invalid qid value", http.StatusBadRequest)
		return
	}

	err = s.repo.Delete(int64(qid))
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
	w.Header().Add("HX-Redirect", "/")
	w.WriteHeader(http.StatusNoContent)
}
