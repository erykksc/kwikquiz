package common

import (
	"time"
)

type Game struct {
	StartedAt time.Time
	EndedAt   time.Time
	Quiz      Quiz
	// Username -> Points
	Points map[string]int
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

type HX_Headers struct {
	HxCurrentURL  string `json:"HX-Current-URL"`
	HxRequest     string `json:"HX-Request"`
	HxTarget      string `json:"HX-Target"`
	HxTrigger     string `json:"HX-Trigger"`
	HxTriggerName string `json:"HX-Trigger-Name"`
}
