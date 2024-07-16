package quiz

import (
	"gorm.io/gorm"
)

type Quiz struct {
	gorm.Model
	ID          uint
	Title       string
	Password    string
	Description string
	Questions   []Question `gorm:"foreignKey:QuizID"`
}

type Question struct {
	gorm.Model
	QuizID  uint
	Text    string
	Answers []Answer `gorm:"foreignKey:QuestionID"`
}

type Answer struct {
	gorm.Model
	QuestionID uint
	IsCorrect  bool
	Text       string
	LaTeX      string
	ImageName  string
	Image      []byte
}

// It is used for faster lookups if only limited data is needed
type QuizMetadata struct {
	ID    uint
	Title string
}
