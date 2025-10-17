package rtc

type WebsocketUpgrader struct {
}

type WebSocketMessage struct {
	Event string      `json:"event,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

type SDPData struct {
	Type string `json:"type"` // "offer" or "answer"
	SDP  string `json:"sdp"`
}

type ICECandidateData struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex uint16 `json:"sdpMLineIndex"`
}
