package quiz

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/erykksc/kwikquiz/internal/common"
	"log"
	"log/slog"
	"net/http"
	"strconv"
)

var QuizzesRepo QuizRepository = NewInMemoryQuizRepository()
var DEBUG = common.DebugOn()

func NewQuizzesRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /quizzes/{$}", getAllQuizzesHandler)
	mux.HandleFunc("GET /quizzes/{qid}", getQuizHandler)
	mux.HandleFunc("POST /quizzes/create/{$}", postQuizHandler)
	mux.HandleFunc("GET /quizzes/create/{$}", getQuizCreateHandler)
	mux.HandleFunc("GET /quizzes/update/{qid}", getQuizUpdateHandler)
	mux.HandleFunc("PUT /quizzes/update/{qid}", updateQuizHandler)
	mux.HandleFunc("DELETE /quizzes/delete/{qid}", deleteQuizHandler)

	// Add quiz if in debug mode
	if common.DevMode() {
		exampleQuiz := Quiz{
			Title:       "Geography",
			Description: "This is a quiz about capitals around the world",
			Questions: []*Question{
				{
					Text: "What is the capital of France?",
					Answers: []*Answer{
						{Text: "Paris", IsCorrect: true},
						{Text: "Berlin", IsCorrect: false},
						{Text: "Warsaw", IsCorrect: false},
						{Text: "Barcelona", IsCorrect: false},
					},
				},
				{
					Text: "On which continent is Russia?",
					Answers: []*Answer{
						{Text: "Europe", IsCorrect: true},
						{Text: "Asia", IsCorrect: true},
						{Text: "North America", IsCorrect: false},
						{Text: "South America", IsCorrect: false},
					},
				},
			},
		}
		QuizzesRepo.AddQuiz(&exampleQuiz)
		exampleQuiz2 := Quiz{
			Title:       "Math",
			Description: "This is a quiz about math",
			Questions: []*Question{
				{
					Text: "What is 2 + 2?",
					Answers: []*Answer{
						{Text: "4", IsCorrect: true},
						{Text: "5", IsCorrect: false},
					},
				},
				{
					Text: "What is 3 * 3?",
					Answers: []*Answer{
						{Text: "9", IsCorrect: true},
						{Text: "6", IsCorrect: false},
					},
				},
			},
		}
		QuizzesRepo.AddQuiz(&exampleQuiz2)
	}

	return mux
}

// TODO: Make it only accessible by admin
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

	quiz, err := QuizzesRepo.GetQuiz(qid)
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
	fmt.Println(quiz)
}

type createQuizForm struct {
	Qid           int
	Title         string
	Description   string
	QuestionOrder string
	Questions     []*Question
	FormError     string
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

func parseQuizForm(r *http.Request) (*Quiz, error) {
	title := r.FormValue("title")
	password := r.FormValue("password")
	description := r.FormValue("description")
	questionOrder := r.FormValue("question-order")

	questions, err := parseQuestions(r)
	if err != nil {
		return &Quiz{}, err
	}

	return &Quiz{
		Title:         title,
		Password:      password,
		Description:   description,
		QuestionOrder: questionOrder,
		Questions:     questions,
	}, nil
}

func parseQuestions(r *http.Request) ([]*Question, error) {
	var questions []*Question
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
		var answers []*Answer
		for answerIndex := 1; answerIndex <= 4; answerIndex++ {
			answerText := r.FormValue("answer-" + strconv.Itoa(questionIndex) + "-" + strconv.Itoa(answerIndex))
			if answerText == "" {
				return nil, fmt.Errorf("missing answer text for question %d, answer %d", questionIndex, answerIndex)
			}
			answers = append(answers, &Answer{
				Number:    answerIndex,
				IsCorrect: answerIndex == correctAnswer,
				Text:      answerText,
			})

		}
		// Append questions to a slice
		questions = append(questions, &Question{
			Number:        questionIndex,
			Text:          questionText,
			Answers:       answers,
			CorrectAnswer: correctAnswer,
		})
		questionIndex++
	}
	return questions, nil
}

func renderQuizCreateForm(w http.ResponseWriter, quiz *Quiz, err error) {
	err = QuizCreateTemplate.ExecuteTemplate(w, "create-form", createQuizForm{
		Title:         quiz.Title,
		Description:   quiz.Description,
		QuestionOrder: quiz.QuestionOrder,
		Questions:     quiz.Questions,
		FormError:     err.Error(),
	})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func redirectToQuiz(w http.ResponseWriter, quizID int) {
	w.Header().Add("HX-Redirect", fmt.Sprintf("/quizzes/%d", quizID))
	w.WriteHeader(http.StatusCreated)
}

func getQuizCreateHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "method", r.Method, "path", r.URL.Path)
	err := QuizCreateTemplate.Execute(w, nil)
	if err != nil {
		log.Printf("Error executing template: %v", err)
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

	quiz, err := QuizzesRepo.GetQuiz(qid)
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
	fmt.Println(quiz)

	// Serialize quiz data to JSON
	quizJSON, err := json.Marshal(quiz)
	if err != nil {
		log.Println("Error marshaling quiz data to JSON:", err)
		http.Error(w, "Error processing quiz data", http.StatusInternalServerError)
		return
	}

	err = QuizUpdateTemplate.Execute(w, map[string]interface{}{
		"QuizJSON": string(quizJSON),
	})
	if err != nil {
		log.Println("Error executing templates:", err)
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

	err = QuizzesRepo.DeleteQuiz(qid)
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
	fmt.Printf("Quiz %d deleted\n", qid)
	w.Header().Add("HX-Redirect", fmt.Sprintf("/"))
	w.WriteHeader(http.StatusNoContent)
}
