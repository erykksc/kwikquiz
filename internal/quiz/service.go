package quiz

type Service struct {
	repo QuizRepository
}

func NewService(repo QuizRepository) Service {
	return Service{
		repo: repo,
	}
}
