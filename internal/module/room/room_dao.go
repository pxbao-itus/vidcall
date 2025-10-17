package room

import "time"

var expandedRoomDuration = 10 * time.Minute // 5 minutes in seconds
type Room struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Users       [2]*string            `json:"users"`
	CreatedAt   int64                 `json:"created_at"`
	CreatedBy   string                `json:"created_by"` // as UserID
	ExpiredAt   *int64                `json:"expired_at"`
	Subscribers map[string]chan Event `json:"-"`
}

func (room Room) Id() string {
	return room.ID
}

func (room Room) IsFull() bool {
	return room.Users[0] != nil && room.Users[1] != nil
}

func (room Room) GetUserDest(userID string) string {
	if room.Users[0] != nil && *room.Users[0] != userID {
		return *room.Users[0]
	}
	if room.Users[1] != nil && *room.Users[1] != userID {
		return *room.Users[1]
	}
	return ""
}

func (room Room) IsExpired() bool {
	return room.ExpiredAt != nil && *room.ExpiredAt < time.Now().Unix()
}

func (room Room) ShouldDelete() bool {
	// Consider room expired if ExpiredAt is set and is older than expandedRoomDuration ago
	return room.ExpiredAt != nil && *room.ExpiredAt < time.Now().Add(-expandedRoomDuration).Unix()
}
