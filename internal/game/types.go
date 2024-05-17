package game

import "time"

type Lobby struct {
	Pin string
	Game
	CreatedAt              time.Time
	CurrentQuestion        Question
	CurrentQuestionTimeout int64 // timestamp (when the server should not accept answers anymore for the current question, the host can send a request to shorten the answer time)
}

type Game struct {
	Hostname        string
	TimePerQuestion time.Duration // time per question
	StartedAt       int64
	EndedAt         int64
}

type Question struct {
	text string
}
