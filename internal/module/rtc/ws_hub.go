package rtc

import "sync"

type WebsocketHub struct {
	clients sync.Map
}
