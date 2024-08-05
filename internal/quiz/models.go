package quiz

import (
	"errors"
	"strconv"

	"github.com/erykksc/kwikquiz/internal/game"
)

type Quiz struct {
	ID          int64  `db:"quiz_id"`
	TitleField  string `db:"title"`
	Password    string
	Description string
	Questions   []Question
}

func (q Quiz) Title() string {
	return q.TitleField
}

func (q Quiz) GetQuestion(idx int) (game.Question, error) {
	if len(q.Questions) > idx && idx > -1 {
		return q.Questions[idx], nil
	}

	return Question{}, errors.New("No question with index: " + strconv.Itoa(idx))
}

func (q Quiz) QuestionsCount() int {
	return len(q.Questions)
}

type Question struct {
	id      int64  `db:"question_id"`
	QuizID  int64  `db:"quiz_id"`
	Text    string `db:"question_text"`
	answers []Answer
}

func (q Question) IsAnswerCorrect(answerIndex int) bool {
	isValid := q.IsAnswerValid(answerIndex)

	if !isValid {
		return false
	}

	return q.answers[answerIndex].IsCorrect
}

func (q Question) IsAnswerValid(answerIndex int) bool {
	if len(q.answers) <= answerIndex {
		return false
	}
	if 0 > answerIndex {
		return false
	}

	return true
}

func (q Question) Answers() []game.Answer {
	answers := make([]game.Answer, len(q.answers))

	for i, answer := range q.answers {
		answers[i] = answer
	}

	return answers
}

type Answer struct {
	ID         int64  `db:"answer_id"`
	QuestionID int64  `db:"question_id"`
	IsCorrect  bool   `db:"is_correct"`
	TextField  string `db:"answer_text"`
	LaTeX      string `db:"latex"`
	ImageName  string
	Image      []byte
}

func (a Answer) Text() string {
	return a.TextField
}

// It is used for faster lookups if only limited data is needed
type QuizMetadata struct {
	ID    uint `db:"quiz_id"`
	Title string
}
