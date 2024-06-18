package pastgames

import "time"

type PlayerScore struct {
	Username string
	Score    int
}

type PastGame struct {
	ID        int
	StartedAt time.Time
	EndedAt   time.Time
	Quiz      string
	Scores    []PlayerScore // sorted by score, descending
}
