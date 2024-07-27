package common

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"net/http"
	"time"
)

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

// ensureClientID returns the clientID from the request cookie, and sets it if non existent
func EnsureClientID(w http.ResponseWriter, r *http.Request) (ClientID, error) {
	// GET CLIENT ID from COOKIE
	cookieCID, err := r.Cookie("client-id")

	// Cookie found
	if err == nil {
		return ClientID(cookieCID.Value), nil
	}

	// Error getting the cookie
	if err != http.ErrNoCookie {
		return "", err
	}

	// Create new cookie
	cID, err := NewClientID()
	if err != nil {
		return "", errors.New("Error generating new client id: " + err.Error())
	}

	// SET CLIENT ID COOKIE or UPDATE EXPIRATION
	http.SetCookie(w, &http.Cookie{
		Name:  "client-id",
		Value: string(cID),
	})

	return cID, nil
}
