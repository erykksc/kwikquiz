package pastgames

import (
	"time"
)

type PastGame struct {
	ID        int64
	StartedAt time.Time     `db:"started_at"`
	EndedAt   time.Time     `db:"ended_at"`
	QuizTitle string        `db:"quiz_title"`
	Scores    []PlayerScore // sorted by score, descending
}

type PlayerScore struct {
	Username string `db:"username"`
	Score    int    `db:"score"`
}
