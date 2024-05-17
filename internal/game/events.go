package game

import "sync"

type GameEvent interface {
	Name() string
}

type GameEventBroadcaster struct {
	mu    sync.Mutex
	chans []chan GameEvent // subscriber channels
}

func (b *GameEventBroadcaster) Subscribe() chan GameEvent {
	ch := make(chan GameEvent)
	b.mu.Lock()
	b.chans = append(b.chans, ch)
	b.mu.Unlock()
	return ch
}

func (b *GameEventBroadcaster) Broadcast(event GameEvent) {
	b.mu.Lock()
	for _, ch := range b.chans {
		ch <- event
	}
	b.mu.Unlock()
}

// GEUserJoined is a game event that is broadcasted when a user joins the game
type GEUserJoined struct {
	Username string
}

func (e GEUserJoined) Name() string {
	return "GEUserJoined"
}

// GEUsernameUpdated is a game event that is broadcasted when a user updates/sets their username
type GEUsernameUpdated struct {
	UserID      int
	NewUsername string
}

func (e GEUsernameUpdated) Name() string {
	return "GEUsernameUpdated"
}
