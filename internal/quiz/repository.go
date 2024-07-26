package quiz

type ErrQuizNotFound struct{}

func (ErrQuizNotFound) Error() string { return "Quiz not found" }

type ErrQuizAlreadyExists struct{}

func (ErrQuizAlreadyExists) Error() string { return "Quiz already exists" }

type Repository interface {
	AddQuiz(Quiz) (uint, error)
	UpdateQuiz(Quiz) (uint, error)
	GetQuiz(id uint) (Quiz, error)
	DeleteQuiz(id uint) error
	GetAllQuizzes() ([]Quiz, error)
	GetAllQuizzesMetadata() ([]QuizMetadata, error)
}
