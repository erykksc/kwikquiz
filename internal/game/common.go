package game

type Username string

func (u Username) IsValid() (reason string, isValid bool) {
	// Check if the username is empty
	if len(u) == 0 {
		return "username is empty", false
	}

	// Check if the username is too long
	if len(u) > 40 {
		return "username is too long", false
	}

	// Check if username contains only empty characters
	for _, c := range u {
		if c != ' ' {
			return "whitespaces not allowed", false
		}
		if c != '\t' {
			return "tabs not allowed", false
		}
	}

	return "", true
}

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
