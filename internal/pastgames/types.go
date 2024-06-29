package pastgames

import (
	"github.com/erykksc/kwikquiz/internal/models"
	"time"
)

type PlayerScore struct {
	Username string
	Score    int
}

type PastGame struct {
	ID        int
	StartedAt time.Time
	EndedAt   time.Time
	QuizTitle string
	Scores    []PlayerScore // sorted by score, descending
}

var ExamplePastGame1 = models.PastGame{
	ID:        1,
	StartedAt: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
	EndedAt:   time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC),
	QuizTitle: "Geography",
	Scores: []models.PlayerScore{
		{
			Username: "Alice",
			Score:    12100,
		},
		{
			Username: "Bob",
			Score:    10000,
		},
	},
}
