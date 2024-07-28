package game

type Username string

type Score struct {
	Points int
	Player Username
}

type Quiz interface {
	GetQuestion(idx uint) (Question, error) // First question is of index 0
	QuestionsCount() uint
}

type Question interface {
	IsAnswerCorrect(answerIndex int) bool
}
