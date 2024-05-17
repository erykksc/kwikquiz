package game

type ErrLobbyNotFound struct{}

func (ErrLobbyNotFound) Error() string {
	return "lobby not found"
}

type ErrLobbyAlreadyExists struct{}

func (ErrLobbyAlreadyExists) Error() string {
	return "game already exists"
}

type LobbyRepository interface {
	AddLobby(lobby Lobby) error
	UpdateLobby(lobby Lobby) error
	GetLobby(pin string) (Lobby, error)
	DeleteLobby(pin string) error
	GetAllLobbies() ([]Lobby, error)
}

// In-memory store for games
type inMemoryLobbyRepository struct {
	lobbies map[string]Lobby
}

func NewInMemoryLobbyRepository() *inMemoryLobbyRepository {
	return &inMemoryLobbyRepository{
		lobbies: make(map[string]Lobby),
	}
}

func (s *inMemoryLobbyRepository) AddLobby(lobby Lobby) error {
	if _, ok := s.lobbies[lobby.Pin]; ok {
		return ErrLobbyAlreadyExists{}
	}

	s.lobbies[lobby.Pin] = lobby
	return nil
}

func (s *inMemoryLobbyRepository) UpdateLobby(lobby Lobby) error {
	if _, ok := s.lobbies[lobby.Pin]; !ok {
		return ErrLobbyNotFound{}
	}

	s.lobbies[lobby.Pin] = lobby
	return nil
}

func (s *inMemoryLobbyRepository) GetLobby(pin string) (Lobby, error) {
	lobby, ok := s.lobbies[pin]
	if !ok {
		return Lobby{}, ErrLobbyNotFound{}
	}

	return lobby, nil
}

func (s *inMemoryLobbyRepository) DeleteLobby(pin string) error {
	if _, ok := s.lobbies[pin]; !ok {
		return ErrLobbyNotFound{}
	}

	delete(s.lobbies, pin)
	return nil
}

func (s *inMemoryLobbyRepository) GetAllLobbies() ([]Lobby, error) {
	var lobbies []Lobby
	for _, lobby := range s.lobbies {
		lobbies = append(lobbies, lobby)
	}

	return lobbies, nil
}
