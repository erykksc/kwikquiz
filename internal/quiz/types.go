package quiz

type Quiz struct {
	ID            int
	Title         string
	Password      string
	Description   string
	QuestionOrder string
	Questions     []Question
}

type Question struct {
	Number        int
	Text          string
	Answers       []Answer
	CorrectAnswer int
}

type Answer struct {
	Number    int
	IsCorrect bool
	Text      string
	// later we can add img, video etc. to allow multimodal questions
}

// It is used for faster lookups if only limited data is needed
type QuizMetadata struct {
	ID    int
	Title string
}
