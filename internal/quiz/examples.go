package quiz

func GetExamples() []Quiz {
	return []Quiz{
		ExampleQuizGeography,
		ExampleQuizMath,
	}
}

var ExampleQuizGeography = Quiz{
	ID:          999,
	TitleField:  "Geography",
	Description: "This is a quiz about capitals around the world",
	Questions: []Question{
		{
			Text: "What is the capital of France?",
			answers: []Answer{
				{TextField: "Paris", IsCorrect: true},
				{TextField: "Berlin", IsCorrect: false},
				{TextField: "Warsaw", IsCorrect: false},
				{TextField: "Barcelona", IsCorrect: false},
			},
		},
		{
			Text: "On which continent is Russia?",
			answers: []Answer{
				{TextField: "Europe", IsCorrect: true},
				{TextField: "Asia", IsCorrect: true},
				{TextField: "North America", IsCorrect: false},
				{TextField: "South America", IsCorrect: false},
			},
		},
	},
}

var ExampleQuizMath = Quiz{
	ID:          998,
	TitleField:  "Math",
	Description: "This is a quiz about math",
	Questions: []Question{
		{
			Text: "What is 2 + 2?",
			answers: []Answer{
				{TextField: "4", IsCorrect: true},
				{TextField: "5", IsCorrect: false},
			},
		},
		{
			Text: "What is 3 * 3?",
			answers: []Answer{
				{TextField: "9", IsCorrect: true},
				{TextField: "6", IsCorrect: false},
			},
		},
	},
}
