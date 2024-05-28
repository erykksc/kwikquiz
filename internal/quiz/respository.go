package quiz

import "sync"

type Quiz struct {
	ID              int
	Title           string
	Password        string
	Description     string
	TimePerQuestion int
	QuestionOrder   string
	Questions       []Question
}

type Question struct {
	Text          string
	Answers       []string
	CorrectAnswer int
}

type ErrQuizNotFound struct{}

func (ErrQuizNotFound) Error() string { return "Quiz not found" }

type ErrQuizAlreadyExists struct{}

func (ErrQuizAlreadyExists) Error() string { return "Quiz already exists" }

type QuizRepository interface {
	AddQuiz(quiz Quiz) (int, error)
	UpdateQuiz(quiz Quiz) error
	GetQuiz(id int) (Quiz, error)
	DeleteQuiz(id int) error
	GetAllQuizzes() ([]Quiz, error)
}

// In-mem store for quizzes
type InMemoryQuizRepository struct {
	quizzes map[int]Quiz
	mutex   sync.RWMutex
	counter int
}

func NewInMemoryQuizRepository() *InMemoryQuizRepository {
	return &InMemoryQuizRepository{
		quizzes: make(map[int]Quiz),
		counter: 0,
	}
}

func (s *InMemoryQuizRepository) AddQuiz(quiz Quiz) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	quiz.ID = s.counter
	s.counter++

	if _, ok := s.quizzes[quiz.ID]; ok {
		return 0, ErrQuizAlreadyExists{}
	}

	s.quizzes[quiz.ID] = quiz
	return quiz.ID, nil
}

func (s *InMemoryQuizRepository) UpdateQuiz(quiz Quiz) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.quizzes[quiz.ID]; !ok {
		return ErrQuizNotFound{}
	}

	s.quizzes[quiz.ID] = quiz
	return nil
}

func (s *InMemoryQuizRepository) GetQuiz(id int) (Quiz, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	quiz, ok := s.quizzes[id]
	if !ok {
		return Quiz{}, ErrQuizNotFound{}
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

func (s *InMemoryQuizRepository) GetAllQuizzes() ([]Quiz, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var quizzes []Quiz
	for _, quiz := range s.quizzes {
		quizzes = append(quizzes, quiz)
	}
	return quizzes, nil
}

func (q *Quiz) AddQuestion(question Question) {
	q.Questions = append(q.Questions, question)
}
