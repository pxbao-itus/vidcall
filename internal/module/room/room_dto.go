package room

const (
	EventNewComer    = "new_comer"
	EventLeaveRoom   = "leave_room"
	EventRoomDeleted = "room_deleted"
)

type Event struct {
	EventName string `json:"event_name,omitempty"`
	Data      any    `json:"data,omitempty"`
}

func (event Event) NewComer() (string, bool) {
	if event.EventName != EventNewComer {
		return "", false
	}
	userID, ok := event.Data.(string)
	if !ok {
		return "", false
	}
	return userID, true
}

func (event Event) LeaveRoom() (string, bool) {
	if event.EventName != EventLeaveRoom {
		return "", false
	}
	userID, ok := event.Data.(string)
	if !ok {
		return "", false
	}
	return userID, true
}

type ListRoomRequest struct {
	OwnerID string `query:"owner_id"`
}
