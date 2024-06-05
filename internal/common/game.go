package common

import "sort"

type Player struct {
	Username string
	Score    int
}

func (g *Game) GetLeaderboard() []*Player {
	points := g.Points
	leaderboard := make([]*Player, 0, len(points))
	for username, score := range points {
		leaderboard = append(leaderboard, &Player{
			Username: username,
			Score:    score,
		})
	}

	sort.Slice(leaderboard, func(i, j int) bool {
		return leaderboard[i].Score > leaderboard[j].Score
	})

	if len(leaderboard) > 3 {
		return leaderboard[:3]
	}
	return leaderboard
}