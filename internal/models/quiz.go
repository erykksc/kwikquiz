package models

import (
	"gorm.io/gorm"
)

type Quiz struct {
	gorm.Model
	ID            uint
	Title         string
	Password      string
	Description   string
	QuestionOrder string
	Questions     []Question `gorm:"foreignKey:ID"`
}

type Question struct {
	gorm.Model
	Text          string
	Answers       []Answer `gorm:"foreignKey:Text"`
	CorrectAnswer int
}

type Answer struct {
	gorm.Model
	IsCorrect bool
	Text      string
}

// It is used for faster lookups if only limited data is needed
type QuizMetadata struct {
	ID    uint
	Title string
}
