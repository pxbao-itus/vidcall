package user

import (
	"time"

	"github.com/gorilla/websocket"
)

type User struct {
	ID         string          `json:"id"`
	Conn       *websocket.Conn `json:"-"`
	LastActive time.Time       `json:"last_active,omitzero"`
}

func (user User) Id() string {
	return user.ID
}

func (user User) IsOnline() bool {
	return user.Conn == nil || time.Since(user.LastActive) > 5*time.Minute
}
