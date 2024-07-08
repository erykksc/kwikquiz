package quiz

import "github.com/erykksc/kwikquiz/internal/models"

var ExampleQuizGeography = models.Quiz{
	ID:          999,
	Title:       "Geography",
	Description: "This is a quiz about capitals around the world",
	Questions: []models.Question{
		{
			Text: "What is the capital of France?",
			Answers: []models.Answer{
				{Text: "Paris", IsCorrect: true},
				{Text: "Berlin", IsCorrect: false},
				{Text: "Warsaw", IsCorrect: false},
				{Text: "Barcelona", IsCorrect: false},
			},
		},
		{
			Text: "On which continent is Russia?",
			Answers: []models.Answer{
				{Text: "Europe", IsCorrect: true},
				{Text: "Asia", IsCorrect: true},
				{Text: "North America", IsCorrect: false},
				{Text: "South America", IsCorrect: false},
			},
		},
	},
}

var ExampleQuizMath = models.Quiz{
	ID:          998,
	Title:       "Math",
	Description: "This is a quiz about math",
	Questions: []models.Question{
		{
			Text: "What is 2 + 2?",
			Answers: []models.Answer{
				{Text: "4", IsCorrect: true},
				{Text: "5", IsCorrect: false},
			},
		},
		{
			Text: "What is 3 * 3?",
			Answers: []models.Answer{
				{Text: "9", IsCorrect: true},
				{Text: "6", IsCorrect: false},
			},
		},
	},
}
