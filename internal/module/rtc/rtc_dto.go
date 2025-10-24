package rtc

type WebsocketUpgrader struct {
}

type WebSocketMessage struct {
	Event string      `json:"event,omitempty"`
	Data  interface{} `json:"data,omitempty"`
	From  string      `json:"from,omitempty"`
	To    string      `json:"to,omitempty"`
}

type RTCSDP struct {
	Type string `json:"type"` // offer or answer
	Sdp  string `json:"sdp"`
}

type RTCIceCandidate struct {
	Candidate     string  `json:"candidate"`
	SdpMid        *string `json:"sdpMid,omitempty"`
	SdpMLineIndex *uint16 `json:"sdpMLineIndex,omitempty"`
}

type PeerInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name,omitempty"`
}

type ChatMessage struct {
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}
