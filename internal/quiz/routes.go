package quiz

import (
	"bytes"
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

	// HTMX question/answer CRUD endpoints for create form
	mux.HandleFunc("POST /quizzes/create/add-question", s.addQuestionCreateHandler)
	mux.HandleFunc("POST /quizzes/create/delete-question/{qidx}", s.deleteQuestionCreateHandler)
	mux.HandleFunc("POST /quizzes/create/add-answer/{qidx}", s.addAnswerCreateHandler)
	mux.HandleFunc("POST /quizzes/create/delete-answer/{qidx}/{aidx}", s.deleteAnswerCreateHandler)

	// HTMX question/answer CRUD endpoints for update form
	mux.HandleFunc("POST /quizzes/update/{qid}/add-question", s.addQuestionUpdateHandler)
	mux.HandleFunc("POST /quizzes/update/{qid}/delete-question/{qidx}", s.deleteQuestionUpdateHandler)
	mux.HandleFunc("POST /quizzes/update/{qid}/add-answer/{qidx}", s.addAnswerUpdateHandler)
	mux.HandleFunc("POST /quizzes/update/{qid}/delete-answer/{qidx}/{aidx}", s.deleteAnswerUpdateHandler)

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

	qid, err := strconv.Atoi(qidStr)
	if err != nil {
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

type createFormData struct {
	LobbyPin     string
	ActionPrefix string
	Title        string
	Password     string
	Description  string
	FormError    string
	Questions    []Question
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

	questions, err := s.parseQuestions(r)
	if err != nil {
		return Quiz{}, err
	}

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

			_, _, err := r.FormFile(answerPrefix)
			textValue := r.FormValue(answerPrefix)

			if err != nil && textValue == "" {
				break
			}

			var answer Answer

			if textValue != "" {
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

			// Read checkbox values for correct answers
			correctAnswers := r.Form["correct-"+strconv.Itoa(questionIndex)]
			for _, correctIdx := range correctAnswers {
				if correctIdx == strconv.Itoa(answerIndex) {
					answer.IsCorrect = true
					break
				}
			}

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
	data := createFormData{
		Title:       quiz.TitleField,
		Password:    quiz.Password,
		Description: quiz.Description,
		Questions:   quiz.Questions,
		FormError:   err.Error(),
		ActionPrefix: "/quizzes/create",
	}
	err = QuizCreateTemplate.ExecuteTemplate(w, "create-form", data)
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

	queryParams := r.URL.Query()
	lobbyPin := queryParams.Get("LobbyPin")

	data := createFormData{
		LobbyPin:     lobbyPin,
		ActionPrefix: "/quizzes/create",
		Questions:    []Question{},
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

	qid, err := strconv.Atoi(qidStr)
	if err != nil {
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

	queryParams := r.URL.Query()
	lobbyPin := queryParams.Get("LobbyPin")

	data := struct {
		Pin string
	}{
		Pin: lobbyPin,
	}

	err = QuizUpdateTemplate.Execute(w, map[string]interface{}{
		"Quiz":         quiz,
		"LobbyPin":     data,
		"ActionPrefix": "/quizzes/update/" + qidStr,
		"Questions":    quiz.Questions,
		"Title":        quiz.TitleField,
		"Password":     quiz.Password,
		"Description":  quiz.Description,
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

	qid, err := strconv.Atoi(qidStr)
	if err != nil {
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

// ---------------------------------------------------------------------------
// HTMX question/answer CRUD handlers
// ---------------------------------------------------------------------------
// All these handlers receive the full form data via HTMX, parse out the
// existing questions, mutate them, then re-render the #questions-list partial.
// ---------------------------------------------------------------------------

func (s Service) addQuestionCreateHandler(w http.ResponseWriter, r *http.Request) {
	questions := s.mutateQuestions(r, func(qs []Question) []Question {
		return append(qs, Question{
			Text:    "",
			answers: []Answer{{TextField: ""}},
		})
	})
	s.renderQuestionList(w, questions, "/quizzes/create")
}

func (s Service) deleteQuestionCreateHandler(w http.ResponseWriter, r *http.Request) {
	qidx := parseIntPathValue(r, "qidx")
	questions := s.mutateQuestions(r, func(qs []Question) []Question {
		idx := qidx - 1
		if idx >= 0 && idx < len(qs) {
			qs = append(qs[:idx], qs[idx+1:]...)
		}
		return qs
	})
	s.renderQuestionList(w, questions, "/quizzes/create")
}

func (s Service) addAnswerCreateHandler(w http.ResponseWriter, r *http.Request) {
	qidx := parseIntPathValue(r, "qidx")
	questions := s.mutateQuestions(r, func(qs []Question) []Question {
		idx := qidx - 1
		if idx >= 0 && idx < len(qs) {
			qs[idx].answers = append(qs[idx].answers, Answer{TextField: ""})
		}
		return qs
	})
	s.renderQuestionList(w, questions, "/quizzes/create")
}

func (s Service) deleteAnswerCreateHandler(w http.ResponseWriter, r *http.Request) {
	qidx := parseIntPathValue(r, "qidx")
	aidx := parseIntPathValue(r, "aidx")
	questions := s.mutateQuestions(r, func(qs []Question) []Question {
		qIdx := qidx - 1
		aIdx := aidx - 1
		if qIdx >= 0 && qIdx < len(qs) && aIdx >= 0 && aIdx < len(qs[qIdx].answers) {
			qs[qIdx].answers = append(qs[qIdx].answers[:aIdx], qs[qIdx].answers[aIdx+1:]...)
		}
		return qs
	})
	s.renderQuestionList(w, questions, "/quizzes/create")
}

func (s Service) addQuestionUpdateHandler(w http.ResponseWriter, r *http.Request) {
	questions := s.mutateQuestions(r, func(qs []Question) []Question {
		return append(qs, Question{
			Text:    "",
			answers: []Answer{{TextField: ""}},
		})
	})
	s.renderQuestionList(w, questions, "/quizzes/update/"+r.PathValue("qid"))
}

func (s Service) deleteQuestionUpdateHandler(w http.ResponseWriter, r *http.Request) {
	qidx := parseIntPathValue(r, "qidx")
	questions := s.mutateQuestions(r, func(qs []Question) []Question {
		idx := qidx - 1
		if idx >= 0 && idx < len(qs) {
			qs = append(qs[:idx], qs[idx+1:]...)
		}
		return qs
	})
	s.renderQuestionList(w, questions, "/quizzes/update/"+r.PathValue("qid"))
}

func (s Service) addAnswerUpdateHandler(w http.ResponseWriter, r *http.Request) {
	qidx := parseIntPathValue(r, "qidx")
	questions := s.mutateQuestions(r, func(qs []Question) []Question {
		idx := qidx - 1
		if idx >= 0 && idx < len(qs) {
			qs[idx].answers = append(qs[idx].answers, Answer{TextField: ""})
		}
		return qs
	})
	s.renderQuestionList(w, questions, "/quizzes/update/"+r.PathValue("qid"))
}

func (s Service) deleteAnswerUpdateHandler(w http.ResponseWriter, r *http.Request) {
	qidx := parseIntPathValue(r, "qidx")
	aidx := parseIntPathValue(r, "aidx")
	questions := s.mutateQuestions(r, func(qs []Question) []Question {
		qIdx := qidx - 1
		aIdx := aidx - 1
		if qIdx >= 0 && qIdx < len(qs) && aIdx >= 0 && aIdx < len(qs[qIdx].answers) {
			qs[qIdx].answers = append(qs[qIdx].answers[:aIdx], qs[qIdx].answers[aIdx+1:]...)
		}
		return qs
	})
	s.renderQuestionList(w, questions, "/quizzes/update/"+r.PathValue("qid"))
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// mutateQuestions parses the form data, applies a mutation function to the
// questions slice, and returns the result.
func (s Service) mutateQuestions(r *http.Request, mutate func([]Question) []Question) []Question {
	questions, err := s.parseQuestions(r)
	if err != nil {
		slog.Error("Error parsing questions from form", "err", err)
		return []Question{}
	}
	return mutate(questions)
}

func (s Service) renderQuestionList(w http.ResponseWriter, questions []Question, actionPrefix string) {
	err := QuestionListTmpl.ExecuteTemplate(w, "question-list", QuestionListData{
		Questions:    questions,
		ActionPrefix: actionPrefix,
	})
	if err != nil {
		slog.Error("Error rendering question list", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func parseIntPathValue(r *http.Request, name string) int {
	val := r.PathValue(name)
	n, err := strconv.Atoi(val)
	if err != nil {
		slog.Error("Invalid path value", "name", name, "value", val)
		return 0
	}
	return n
}
