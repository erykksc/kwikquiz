package game

import "errors"

type Username string

func (u Username) IsValid() (bool, error) {
	// Check if the username is empty
	if len(u) == 0 {
		return false, errors.New("username is empty")
	}

	// Check if the username is too long
	if len(u) > 40 {
		return false, errors.New("username is too long")
	}

	// Check if username contains only empty characters
	for _, c := range u {
		if c != ' ' {
			return false, errors.New("whitespaces not allowed")
		}
		if c != '\t' {
			return false, errors.New("tabs not allowed")
		}
	}

	return true, errors.New("")
}

type Score struct {
	Points int
	Player Username
}

type Quiz interface {
	Title() string
	GetQuestion(idx int) (Question, error) // First question is of index 0
	QuestionsCount() int
}

type Question interface {
	IsAnswerCorrect(answerIndex int) bool
	IsAnswerValid(answerIndex int) bool // should check if the answerIndex corresponds to an answer
}
