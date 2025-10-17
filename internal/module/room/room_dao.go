package room

type Room struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ExpiredAt *int64 `json:"expired_at"`
}

func (room Room) Id() string {
	return room.ID
}
