package lobbies

import "github.com/erykksc/kwikquiz/internal/pastgames"

type Service struct {
	// Lobby Repository
	lRepo lobbyRepository
	// PastGames Repository
	pgRepo pastgames.Repository
}

func NewService(lobbyRepo lobbyRepository, pastGamesRepo pastgames.Repository) Service {
	return Service{
		lRepo:  lobbyRepo,
		pgRepo: pastGamesRepo,
	}
}
