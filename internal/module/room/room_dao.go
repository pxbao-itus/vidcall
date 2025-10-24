package room

import (
	"fmt"
	"time"
)

var expandedRoomDuration = 10 * time.Minute // 5 minutes in seconds
type Room struct {
	ID          string                `json:"id"`
	Name        string                `json:"name,omitempty"`
	Description string                `json:"description,omitempty"`
	Users       []string              `json:"users,omitempty"`
	CreatedAt   int64                 `json:"created_at,omitempty"`
	CreatedBy   string                `json:"created_by,omitempty"` // as UserID
	ExpiredAt   int64                 `json:"expired_at,omitempty"`
	MaxGuest    int                   `json:"max_guest,omitempty"`
	Subscribers map[string]chan Event `json:"-"`
}

func (room Room) Id() string {
	return room.ID
}

func (room Room) IsFull() bool {
	return room.MaxGuest > 0 && len(room.Users) >= room.MaxGuest
}

func (room Room) IsExpired() bool {
	return room.ExpiredAt > 0 && room.ExpiredAt < time.Now().Unix()
}

func (room Room) shouldDelete() bool {
	// Consider room expired if ExpiredAt is set and is older than expandedRoomDuration ago
	return room.ExpiredAt > 0 && room.ExpiredAt < time.Now().Add(-expandedRoomDuration).Unix()
}

func (room *Room) AddGuest(userID string) {
	room.Users = append(room.Users, userID)
	room.Subscribers[userID] = make(chan Event, 5) // Buffered channel to avoid blocking

	newComerEvent := Event{
		EventName: EventNewComer,
		Data:      userID,
	}
	room.emitEvent(userID, newComerEvent)
}

func (room *Room) RemoveGuest(userID string) {
	delete(room.Subscribers, userID)
	// Remove userID from Users slice
	for i, id := range room.Users {
		if id == userID {
			room.Users = append(room.Users[:i], room.Users[i+1:]...)
			break
		}
	}
	leaveEvent := Event{
		EventName: EventLeaveRoom,
		Data:      userID,
	}
	room.emitEvent(userID, leaveEvent)
}

func (room *Room) emitEvent(userID string, event Event) {
	for id, ch := range room.Subscribers {
		if id == userID {
			continue
		}

		//if _, ok := <-ch; !ok {
		//	// If the channel is blocked or closed, skip sending the event
		//	continue
		//}

		go func(ch chan Event) {
			ch <- event
			fmt.Println("Event emitted to user:", id, "event:", event)
		}(ch)
	}
}
