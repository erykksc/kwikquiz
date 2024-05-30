package lobby

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"sync"
	"time"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/gorilla/websocket"
)

type Lobby struct {
	mu              sync.Mutex
	Host            ClientID
	Pin             string
	TimePerQuestion time.Duration // time per question
	common.Game
	CreatedAt              time.Time
	CurrentQuestion        common.Question
	CurrentQuestionTimeout time.Time // timestamp (when the server should not accept answers anymore for the current question, the host can send a request to shorten the answer time)
	Players                map[ClientID]*Player
}

type ClientID string

func NewClientID() (ClientID, error) {
	// Generate 8 bytes from the timestamp (64 bits)
	timestamp := time.Now().Unix()
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))

	// Generate 8 random bytes (64 bits)
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Combine the two byte slices into 128 bits
	combinedBytes := append(timestampBytes, randomBytes...)

	// Encode the 128 bits into a base64 string
	encoded := base64.StdEncoding.EncodeToString(combinedBytes)
	return ClientID(encoded), nil
}

type Player struct {
	Conn     *websocket.Conn
	Username string
	Score 	 int
}

type LobbyOptions struct {
	TimePerQuestion time.Duration
	Pin             string
}

func CreateLobby(options LobbyOptions) *Lobby {
	return &Lobby{
		TimePerQuestion: options.TimePerQuestion,
		Pin:             options.Pin,
		Game:            common.Game{},
		CreatedAt:       time.Now(),
		Players:         make(map[ClientID]*Player),
	}
}

//returns top 3 players in a lobby
func (l *Lobby) GetLeaderboard() []*Player {
	l.mu.Lock()
	defer l.mu.Unlock()

	players := make([]*Player, 0, len(l.Players))
	for _, player := range l.Players {
		players = append(players, player)
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].Score > players[j].Score
	})

	if len(players) > 3 {
		return players[:3]
	}
	return players
}
