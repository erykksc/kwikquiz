package quiz

import (
	"sync"
)

type Quiz struct {
	ID            int
	Title         string
	Password      string
	Description   string
	QuestionOrder string
	Questions     []*Question
	mutex         sync.RWMutex
}

type Question struct {
	Number        int
	Text          string
	Answers       []*Answer
	CorrectAnswer int
	mutex         sync.RWMutex
}

type Answer struct {
	Number    int
	IsCorrect bool
	Text      string
	// later we can add img, video etc. to allow multimodal questions
}

type ErrQuizNotFound struct{}

func (ErrQuizNotFound) Error() string { return "Quiz not found" }

type ErrQuizAlreadyExists struct{}

func (ErrQuizAlreadyExists) Error() string { return "Quiz already exists" }

type QuizRepository interface {
	AddQuiz(*Quiz) (int, error)
	UpdateQuiz(*Quiz) (int, error)
	GetQuiz(id int) (*Quiz, error)
	DeleteQuiz(id int) error
	GetAllQuizzes() ([]*Quiz, error)
}

// InMemoryQuizRepository In-mem store for quizzes
type InMemoryQuizRepository struct {
	quizzes map[int]*Quiz
	mutex   sync.RWMutex
	counter int
}

func NewInMemoryQuizRepository() *InMemoryQuizRepository {
	return &InMemoryQuizRepository{
		quizzes: make(map[int]*Quiz),
		counter: 0,
	}
}

func (s *InMemoryQuizRepository) AddQuiz(q *Quiz) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	q.ID = s.counter
	s.counter++

	if _, ok := s.quizzes[q.ID]; ok {
		return 0, ErrQuizAlreadyExists{}
	}

	s.quizzes[q.ID] = q
	return q.ID, nil
}

func (s *InMemoryQuizRepository) UpdateQuiz(q *Quiz) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.quizzes[q.ID]; !ok {
		return 0, ErrQuizNotFound{}
	}

	s.quizzes[q.ID] = q
	return q.ID, nil
}

func (s *InMemoryQuizRepository) GetQuiz(id int) (*Quiz, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	quiz, ok := s.quizzes[id]
	if !ok {
		return &Quiz{}, ErrQuizNotFound{}
	}

	return quiz, nil
}

func (s *InMemoryQuizRepository) DeleteQuiz(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.quizzes[id]; !ok {
		return ErrQuizNotFound{}
	}

	delete(s.quizzes, id)
	return nil
}

func (s *InMemoryQuizRepository) GetAllQuizzes() ([]*Quiz, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var quizzes []*Quiz
	for _, quiz := range s.quizzes {
		quizzes = append(quizzes, quiz)
	}
	return quizzes, nil
}

func (q *Quiz) AddQuestion(question *Question) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	q.Questions = append(q.Questions, question)
}

func (q *Quiz) GetAllQuestions() ([]*Question, error) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	var questions []*Question

	for _, question := range q.Questions {
		questions = append(questions, question)
	}
	return questions, nil
}
