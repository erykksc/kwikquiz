package quiz

type ErrQuizNotFound struct{}

func (ErrQuizNotFound) Error() string { return "Quiz not found" }

type ErrQuizAlreadyExists struct{}

func (ErrQuizAlreadyExists) Error() string { return "Quiz already exists" }

type Repository interface {
	Insert(*Quiz) (int64, error)
	Upsert(*Quiz) (int64, error)
	Update(*Quiz) (int64, error)
	Get(id int64) (*Quiz, error)
	Delete(id int64) error
	GetAll() ([]Quiz, error)
	GetAllQuizzesMetadata() ([]QuizMetadata, error)
}
