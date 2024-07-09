package pastgames

import (
	"github.com/erykksc/kwikquiz/internal/models"
	"time"
)

var ExamplePastGame1 = models.PastGame{
	ID:        999,
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
