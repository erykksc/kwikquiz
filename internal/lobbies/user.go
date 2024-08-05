package lobbies

import (
	"errors"
	"html/template"
	"log/slog"

	"github.com/erykksc/kwikquiz/internal/common"
	"github.com/erykksc/kwikquiz/internal/game"
	"github.com/gorilla/websocket"
)

type User struct {
	Conn     *websocket.Conn
	ClientID common.ClientID
	Username game.Username
}

// writeTemplate does tmpl.Execute(w, data) on websocket connection to the user
func (client *User) writeTemplate(tmpl *template.Template, data any) error {
	slog.Debug("writeTemplate", "template", tmpl.Name(), "client-id", client.ClientID)
	if client.Conn == nil {
		return errors.New("client.Conn is nil")
	}
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
func (client *User) writeNamedTemplate(tmpl *template.Template, name string, data any) error {
	if client.Conn == nil {
		return errors.New("client.Conn is nil")
	}
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
