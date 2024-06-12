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
	Questions     []Question
}

type Question struct {
	Number        int
	Text          string
	Answers       []Answer
	CorrectAnswer int
}

type Answer struct {
	Number    int
	IsCorrect bool
	Text      string
	// later we can add img, video etc. to allow multimodal questions
}

// It is used for faster lookups if only limited data is needed
type QuizMetadata struct {
	ID    int
	Title string
}

type ErrQuizNotFound struct{}

func (ErrQuizNotFound) Error() string { return "Quiz not found" }

type ErrQuizAlreadyExists struct{}

func (ErrQuizAlreadyExists) Error() string { return "Quiz already exists" }

type QuizRepository interface {
	AddQuiz(Quiz) (int, error)
	UpdateQuiz(Quiz) (int, error)
	GetQuiz(id int) (Quiz, error)
	DeleteQuiz(id int) error
	GetAllQuizzes() ([]Quiz, error)
	GetAllQuizzesMetadata() ([]QuizMetadata, error)
}

// InMemoryQuizRepository In-mem store for quizzes
type InMemoryQuizRepository struct {
	quizzes   map[int]Quiz
	mutex     sync.RWMutex
	highestID int
}

func NewInMemoryQuizRepository() *InMemoryQuizRepository {
	return &InMemoryQuizRepository{
		quizzes:   make(map[int]Quiz),
		highestID: 0,
	}
}

func (s *InMemoryQuizRepository) AddQuiz(q Quiz) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if q.ID == 0 {
		// Assign a unique ID
		s.highestID++
		q.ID = s.highestID
	} else if q.ID > s.highestID {
		s.highestID = q.ID
	}

	if _, ok := s.quizzes[q.ID]; ok {
		return 0, ErrQuizAlreadyExists{}
	}

	s.quizzes[q.ID] = q
	return q.ID, nil
}

func (s *InMemoryQuizRepository) UpdateQuiz(q Quiz) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.quizzes[q.ID]; !ok {
		return 0, ErrQuizNotFound{}
	}

	s.quizzes[q.ID] = q
	return q.ID, nil
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

func (s *InMemoryQuizRepository) GetAllQuizzesMetadata() ([]QuizMetadata, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var quizzes []QuizMetadata
	for _, quiz := range s.quizzes {
		quizzes = append(quizzes, QuizMetadata{
			ID:    quiz.ID,
			Title: quiz.Title,
		})
	}
	return quizzes, nil
}
