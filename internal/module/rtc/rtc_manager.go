package rtc

import (
	"context"
	"fmt"
	"sync"

	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
)

type RTCManager struct {
	peerConnections map[string]map[string]map[string]*webrtc.PeerConnection
	mu              sync.RWMutex
	logger          *zap.Logger
	config          webrtc.Configuration
}

func NewRTCManager(logger *zap.Logger) *RTCManager {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{
					"stun:stun.l.google.com:19302",
					"stun:stun1.l.google.com:19302",
				},
			},
		},
	}

	return &RTCManager{
		peerConnections: make(map[string]map[string]map[string]*webrtc.PeerConnection),
		logger:          logger,
		config:          config,
	}
}

func (m *RTCManager) CreatePeerConnection(roomID, userID, peerID string) (*webrtc.PeerConnection, error) {
	pc, err := webrtc.NewPeerConnection(m.config)
	if err != nil {
		return nil, fmt.Errorf("create peer connection: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.peerConnections[roomID] == nil {
		m.peerConnections[roomID] = make(map[string]map[string]*webrtc.PeerConnection)
	}
	if m.peerConnections[roomID][userID] == nil {
		m.peerConnections[roomID][userID] = make(map[string]*webrtc.PeerConnection)
	}

	m.peerConnections[roomID][userID][peerID] = pc

	m.logger.Info("Created peer connection",
		zap.String("roomID", roomID),
		zap.String("userID", userID),
		zap.String("peerID", peerID))

	return pc, nil
}

func (m *RTCManager) GetPeerConnection(roomID, userID, peerID string) *webrtc.PeerConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if room, ok := m.peerConnections[roomID]; ok {
		if user, ok := room[userID]; ok {
			if peer, ok := user[peerID]; ok {
				return peer
			}
		}
	}
	return nil
}

func (m *RTCManager) GetOrCreatePeerConnection(roomID, userID, peerID string) (*webrtc.PeerConnection, error) {
	pc := m.GetPeerConnection(roomID, userID, peerID)
	if pc != nil {
		return pc, nil
	}
	return m.CreatePeerConnection(roomID, userID, peerID)
}

func (m *RTCManager) RemovePeerConnection(roomID, userID, peerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if room, ok := m.peerConnections[roomID]; ok {
		if user, ok := room[userID]; ok {
			if peer, ok := user[peerID]; ok {
				if err := peer.Close(); err != nil {
					m.logger.Error("Close peer connection failed", zap.Error(err))
				}
				delete(user, peerID)

				m.logger.Info("Removed peer connection",
					zap.String("roomID", roomID),
					zap.String("userID", userID),
					zap.String("peerID", peerID))

				if len(user) == 0 {
					delete(room, userID)
				}
			}
		}

		if len(room) == 0 {
			delete(m.peerConnections, roomID)
		}
	}

	return nil
}

func (m *RTCManager) RemoveUserConnections(roomID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if room, ok := m.peerConnections[roomID]; ok {
		if user, ok := room[userID]; ok {
			for peerID, peer := range user {
				if err := peer.Close(); err != nil {
					m.logger.Error("Close peer connection failed",
						zap.String("peerID", peerID),
						zap.Error(err))
				}
			}
			delete(room, userID)

			m.logger.Info("Removed all user connections",
				zap.String("roomID", roomID),
				zap.String("userID", userID))
		}

		if len(room) == 0 {
			delete(m.peerConnections, roomID)
		}
	}

	return nil
}

func (m *RTCManager) HandleOffer(_ context.Context, roomID, userID, peerID string, sdp RTCSDP, onICE func(candidate RTCIceCandidate), onAnswer func(sdp RTCSDP)) error {
	pc, err := m.GetOrCreatePeerConnection(roomID, userID, peerID)
	if err != nil {
		return fmt.Errorf("get or create peer connection: %w", err)
	}

	// Setup ICE candidate handler
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			candidateJSON := candidate.ToJSON()
			onICE(RTCIceCandidate{
				Candidate:     candidateJSON.Candidate,
				SdpMid:        candidateJSON.SDPMid,
				SdpMLineIndex: candidateJSON.SDPMLineIndex,
			})
		}
	})

	// Setup track handler to relay media
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		m.logger.Info("Received track",
			zap.String("userID", userID),
			zap.String("peerID", peerID),
			zap.String("kind", track.Kind().String()),
			zap.String("id", track.ID()),
			zap.String("streamID", track.StreamID()))

		// You can relay this track to other peers if needed
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		m.logger.Info("Peer connection state changed",
			zap.String("userID", userID),
			zap.String("peerID", peerID),
			zap.String("state", state.String()))

		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			_ = m.RemovePeerConnection(roomID, userID, peerID)
		}
	})

	// Set remote description
	if err := pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp.Sdp,
	}); err != nil {
		return fmt.Errorf("set remote description: %w", err)
	}

	// Create answer
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("create answer: %w", err)
	}

	// Set local description
	if err := pc.SetLocalDescription(answer); err != nil {
		return fmt.Errorf("set local description: %w", err)
	}

	onAnswer(RTCSDP{
		Type: "answer",
		Sdp:  answer.SDP,
	})

	return nil
}

func (m *RTCManager) HandleAnswer(roomID, userID, peerID string, sdp RTCSDP) error {
	pc := m.GetPeerConnection(roomID, userID, peerID)
	if pc == nil {
		return fmt.Errorf("peer connection not found for user %s and peer %s", userID, peerID)
	}

	if err := pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp.Sdp,
	}); err != nil {
		return fmt.Errorf("set remote description: %w", err)
	}

	m.logger.Info("Set answer",
		zap.String("roomID", roomID),
		zap.String("userID", userID),
		zap.String("peerID", peerID))

	return nil
}

func (m *RTCManager) HandleICECandidate(roomID, userID, peerID string, candidate RTCIceCandidate) error {
	pc := m.GetPeerConnection(roomID, userID, peerID)
	if pc == nil {
		return fmt.Errorf("peer connection not found for user %s and peer %s", userID, peerID)
	}

	var iceCandidate webrtc.ICECandidateInit
	iceCandidate.Candidate = candidate.Candidate
	if candidate.SdpMid != nil {
		iceCandidate.SDPMid = candidate.SdpMid
	}
	if candidate.SdpMLineIndex != nil {
		iceCandidate.SDPMLineIndex = candidate.SdpMLineIndex
	}

	if err := pc.AddICECandidate(iceCandidate); err != nil {
		return fmt.Errorf("add ICE candidate: %w", err)
	}

	m.logger.Debug("Added ICE candidate",
		zap.String("roomID", roomID),
		zap.String("userID", userID),
		zap.String("peerID", peerID))

	return nil
}

func (m *RTCManager) CreateOffer(roomID, userID, peerID string, onICE func(candidate RTCIceCandidate)) (*RTCSDP, error) {
	pc, err := m.GetOrCreatePeerConnection(roomID, userID, peerID)
	if err != nil {
		return nil, fmt.Errorf("get or create peer connection: %w", err)
	}

	// Setup ICE candidate handler
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			candidateJSON := candidate.ToJSON()
			onICE(RTCIceCandidate{
				Candidate:     candidateJSON.Candidate,
				SdpMid:        candidateJSON.SDPMid,
				SdpMLineIndex: candidateJSON.SDPMLineIndex,
			})
		}
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		m.logger.Info("Peer connection state changed",
			zap.String("userID", userID),
			zap.String("peerID", peerID),
			zap.String("state", state.String()))

		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			_ = m.RemovePeerConnection(roomID, userID, peerID)
		}
	})

	// Add transceiver for video and audio
	if _, err := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		return nil, fmt.Errorf("add video transceiver: %w", err)
	}
	if _, err := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		return nil, fmt.Errorf("add audio transceiver: %w", err)
	}

	// Create offer
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("create offer: %w", err)
	}

	// Set local description
	if err := pc.SetLocalDescription(offer); err != nil {
		return nil, fmt.Errorf("set local description: %w", err)
	}

	m.logger.Info("Created offer",
		zap.String("roomID", roomID),
		zap.String("userID", userID),
		zap.String("peerID", peerID))

	return &RTCSDP{
		Type: "offer",
		Sdp:  offer.SDP,
	}, nil
}
