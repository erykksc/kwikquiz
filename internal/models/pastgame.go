package models

import (
	"gorm.io/gorm"
	"time"
)

type PlayerScore struct {
	gorm.Model
	Username string
	Score    int
}

type PastGame struct {
	gorm.Model
	ID        uint
	StartedAt time.Time
	EndedAt   time.Time
	QuizTitle string
	Scores    []PlayerScore `gorm:"foreignKey:ID"` // sorted by score, descending
}
