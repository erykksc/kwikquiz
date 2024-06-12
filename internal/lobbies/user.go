package lobbies

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"html/template"
	"time"

	"github.com/gorilla/websocket"
)

type user struct {
	Conn                 *websocket.Conn
	ClientID             clientID
	Username             string
	IsHost               bool
	SubmittedAnswerIdx   int
	AnswerSubmissionTime time.Time
	Score                int64
	NewPoints            int64
}

// writeTemplate does tmpl.Execute(w, data) on websocket connection to the user
func (client *user) writeTemplate(tmpl *template.Template, data any) error {
	w, err := client.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	if err := tmpl.Execute(w, data); err != nil {
		return err
	}
	return nil
}

// writeNamedTemplate does tmpl.ExecuteTemplate(w, name, data) on websocket connection to the user
func (client *user) writeNamedTemplate(tmpl *template.Template, name string, data any) error {
	w, err := client.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		return err
	}
	return nil
}

type clientID string

func newClientID() (clientID, error) {
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
	return clientID(encoded), nil
}
