package pastgames

import "time"

func GetExamples() []PastGame {
	return []PastGame{
		ExamplePastGame1,
	}
}

var ExamplePastGame1 = PastGame{
	ID:        999,
	StartedAt: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
	EndedAt:   time.Date(2020, 1, 1, 12, 30, 0, 0, time.UTC),
	QuizTitle: "Geography",
	Scores: []PlayerScore{
		{
			Username: "Alice",
			Score:    12100,
		},
		{
			Username: "Bob",
			Score:    10000,
		},
		{
			Username: "Jack",
			Score:    9000,
		},
		{
			Username: "Max",
			Score:    8500,
		},
		{
			Username: "Jamal",
			Score:    5432,
		},
	},
}
