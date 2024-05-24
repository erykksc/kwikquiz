package common

import (
	"time"
)

type Game struct {
	StartedAt time.Time
	EndedAt   time.Time
	// Username -> Points
	Points map[string]int
	Rounds []Round
}

type Round struct {
	Question        Question
	PossibleAnswers []Answer
	PlayersAnswers  []PlayerAnswer
}

type PlayerAnswer struct {
	Username    string
	SubbmitedAt time.Time
	Answer      Answer
}

type Quiz struct {
	ID          string
	Title       string
	Description string
	Owner       string
	Questions   []Question
}

type Question struct {
	Text    string
	Answers []Answer
	// later we can add img, video etc. to allow multimodal questions
}

type Answer struct {
	IsCorrect bool // multiple answers to a question may be right, can be used in tricky questions
	Text      string
	// later we can add img, video etc. to allow multimodal questions
}
