package rtc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"vidcall/internal/common"
	"vidcall/internal/module/room"
	"vidcall/internal/module/user"

	"github.com/gorilla/websocket"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("rtc",
	fx.Provide(NewRTCManager),
	fx.Provide(NewHandler),
)

type Handler struct {
	roomService *room.Service
	userService *user.Service
	rtcManager  *RTCManager

	logger *zap.Logger
}

type HandlerParams struct {
	fx.In

	RoomService *room.Service
	UserService *user.Service
	RTCManager  *RTCManager
	Logger      *zap.Logger
}

func NewHandler(params HandlerParams) *Handler {
	return &Handler{
		roomService: params.RoomService,
		userService: params.UserService,
		rtcManager:  params.RTCManager,
		logger:      params.Logger,
	}
}

func (handler *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := common.GetParam(r, "roomID")
	if roomID == "" {
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	ws := websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin:       func(r *http.Request) bool { return true },
	}

	conn, err := ws.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, _ := common.GetUserID(r)
	usr := user.User{
		ID:   userID,
		Conn: conn,
	}

	if _, err := handler.userService.CreateOrUpdateUser(r.Context(), usr); err != nil {
		handler.logger.Error("Create user failed", zap.String("userID", userID), zap.Error(err))
		http.Error(w, "Create user failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := conn.Close(); err != nil {
			handler.logger.Error("WebSocket close failed", zap.String("userID", userID), zap.Error(err))
		}

		usr.Conn = nil
		if _, err := handler.userService.UpdateUser(context.TODO(), usr); err != nil {
			handler.logger.Error("Update user conn to nil failed", zap.String("userID", userID), zap.Error(err))
		}
	}()

	commonRoom, err := handler.roomService.JoinRoom(r.Context(), roomID, userID)
	if err != nil {
		handler.logger.Error("Join room failed", zap.String("roomID", roomID), zap.String("userID", userID), zap.Error(err))
		http.Error(w, "Join room failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		// Clean up WebRTC connections
		if err := handler.rtcManager.RemoveUserConnections(commonRoom.ID, userID); err != nil {
			handler.logger.Error("Remove user connections failed",
				zap.String("roomID", commonRoom.ID),
				zap.String("userID", userID),
				zap.Error(err))
		}

		if err := handler.roomService.LeaveRoom(context.TODO(), commonRoom.ID, userID); err != nil {
			handler.logger.Error("Leave room failed",
				zap.String("roomID", commonRoom.ID),
				zap.String("userID", userID),
				zap.Error(err))

			return
		}
	}()

	clientEvent := make(chan any, 5)
	go func() {
		for {
			select {
			case <-r.Context().Done():
				return
			default:
				var msg WebSocketMessage
				if err := usr.Conn.ReadJSON(&msg); err != nil {
					handler.logger.Error("Read msg failed", zap.String("userID", userID), zap.Error(err))
					clientEvent <- fmt.Errorf("read msg: %w", err)
					return
				}

				handler.logger.Info("Received msg", zap.String("userID", userID), zap.Any("msg", msg))
				clientEvent <- msg
			}
		}
	}()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	subscriberChan := commonRoom.Subscribers[userID]

	for {
		select {
		case <-ticker.C:
			if _, err := handler.userService.UpdateActive(r.Context(), userID); err != nil {
				handler.logger.Error("Update user active failed", zap.String("userID", userID), zap.Error(err))
				return
			}
		case <-r.Context().Done():
			return

		case roomEvent := <-subscriberChan:
			switch roomEvent.EventName {
			case room.EventLeaveRoom, room.EventNewComer:
				guestID, ok := roomEvent.Data.(string)
				if !ok {
					handler.logger.Error("Invalid guest ID type", zap.String("roomID", commonRoom.ID), zap.String("userID", userID))
					continue
				}
				usr, err := handler.userService.GetUser(r.Context(), guestID)
				if err != nil {
					handler.logger.Error("Get peer user failed", zap.String("roomID", commonRoom.ID), zap.String("userID", userID), zap.Error(err))
					continue
				}
				msg := WebSocketMessage{
					Event: roomEvent.EventName,
					Data:  usr,
				}
				_ = conn.WriteJSON(msg)
			case room.EventRoomDeleted:
				handler.logger.Info("Room deleted, closing connection", zap.String("roomID", commonRoom.ID), zap.String("userID", userID))
				msg := WebSocketMessage{
					Event: "room_deleted",
				}
				_ = conn.WriteJSON(msg)
				break
			case room.EventNewMsg:
				msgData, ok := roomEvent.Data.(WebSocketMessage)
				if !ok {
					handler.logger.Error("Invalid msg data type", zap.String("roomID", commonRoom.ID), zap.String("userID", userID))
					continue
				}
				_ = conn.WriteJSON(msgData)
			}
		case msg := <-clientEvent:
			if err, ok := msg.(error); ok {
				handler.logger.Error("Read msg failed", zap.String("roomID", commonRoom.ID), zap.String("userID", userID), zap.Error(err))
				return
			}

			clientMsg, ok := msg.(WebSocketMessage)
			if !ok {
				handler.logger.Error("Invalid msg type", zap.String("roomID", commonRoom.ID), zap.String("userID", userID))
				continue
			}

			// Handle WebRTC signaling messages
			switch clientMsg.Event {
			case EventOffer:
				// Forward offer to target peer
				if clientMsg.To != "" {
					handler.sendToUser(r.Context(), clientMsg.To, WebSocketMessage{
						Event: EventOffer,
						From:  userID,
						Data:  clientMsg.Data,
					})
				}
				continue
			case EventAnswer:
				// Forward answer to target peer (the one who sent the offer)
				if clientMsg.To != "" {
					handler.sendToUser(r.Context(), clientMsg.To, WebSocketMessage{
						Event: EventAnswer,
						From:  userID,
						Data:  clientMsg.Data,
					})
				}
				continue
			case EventIceCandidate:
				// Forward ICE candidate to target peer
				if clientMsg.To != "" {
					handler.sendToUser(r.Context(), clientMsg.To, WebSocketMessage{
						Event: EventIceCandidate,
						From:  userID,
						Data:  clientMsg.Data,
					})
				}
				continue
			case EventChatMessage:
				// Broadcast chat message to all users in the room
				roomEvent := room.Event{
					EventName: room.EventNewMsg,
					Data:      clientMsg,
				}
				handler.roomService.EmitEvent(commonRoom.ID, userID, roomEvent)
				continue
			}

			roomEvent := room.Event{
				EventName: room.EventNewMsg,
				Data:      clientMsg,
			}

			handler.roomService.EmitEvent(commonRoom.ID, userID, roomEvent)
		}
	}
}

func (handler *Handler) handleOffer(ctx context.Context, roomID, userID string, msg WebSocketMessage) {
	var sdp RTCSDP
	data, err := json.Marshal(msg.Data)
	if err != nil {
		handler.logger.Error("Marshal offer data failed", zap.Error(err))
		return
	}

	if err := json.Unmarshal(data, &sdp); err != nil {
		handler.logger.Error("Unmarshal offer failed", zap.Error(err))
		return
	}

	peerID := msg.To
	if peerID == "" {
		handler.logger.Error("Peer ID is empty in offer")
		return
	}

	handler.logger.Info("Handling offer",
		zap.String("roomID", roomID),
		zap.String("from", userID),
		zap.String("to", peerID))

	// Handle the offer and generate answer
	err = handler.rtcManager.HandleOffer(ctx, roomID, peerID, userID, sdp,
		func(candidate RTCIceCandidate) {
			// Send ICE candidate to the offerer
			msg := WebSocketMessage{
				Event: EventIceCandidate,
				Data:  candidate,
				From:  peerID,
				To:    userID,
			}
			handler.sendToUser(ctx, userID, msg)
		},
		func(answer RTCSDP) {
			// Send answer back to the offerer
			msg := WebSocketMessage{
				Event: EventAnswer,
				Data:  answer,
				From:  peerID,
				To:    userID,
			}
			handler.sendToUser(ctx, userID, msg)
		})

	if err != nil {
		handler.logger.Error("Handle offer failed",
			zap.String("roomID", roomID),
			zap.String("userID", userID),
			zap.String("peerID", peerID),
			zap.Error(err))
	}
}

func (handler *Handler) handleAnswer(roomID, userID string, msg WebSocketMessage) {
	var sdp RTCSDP
	data, err := json.Marshal(msg.Data)
	if err != nil {
		handler.logger.Error("Marshal answer data failed", zap.Error(err))
		return
	}

	if err := json.Unmarshal(data, &sdp); err != nil {
		handler.logger.Error("Unmarshal answer failed", zap.Error(err))
		return
	}

	peerID := msg.From
	if peerID == "" {
		handler.logger.Error("Peer ID is empty in answer")
		return
	}

	handler.logger.Info("Handling answer",
		zap.String("roomID", roomID),
		zap.String("from", peerID),
		zap.String("to", userID))

	if err := handler.rtcManager.HandleAnswer(roomID, userID, peerID, sdp); err != nil {
		handler.logger.Error("Handle answer failed",
			zap.String("roomID", roomID),
			zap.String("userID", userID),
			zap.String("peerID", peerID),
			zap.Error(err))
	}
}

func (handler *Handler) handleICECandidate(roomID, userID string, msg WebSocketMessage) {
	var candidate RTCIceCandidate
	data, err := json.Marshal(msg.Data)
	if err != nil {
		handler.logger.Error("Marshal ICE candidate data failed", zap.Error(err))
		return
	}

	if err := json.Unmarshal(data, &candidate); err != nil {
		handler.logger.Error("Unmarshal ICE candidate failed", zap.Error(err))
		return
	}

	peerID := msg.From
	if peerID == "" {
		peerID = msg.To
	}

	if peerID == "" {
		handler.logger.Error("Peer ID is empty in ICE candidate")
		return
	}

	handler.logger.Debug("Handling ICE candidate",
		zap.String("roomID", roomID),
		zap.String("userID", userID),
		zap.String("peerID", peerID))

	if err := handler.rtcManager.HandleICECandidate(roomID, userID, peerID, candidate); err != nil {
		handler.logger.Error("Handle ICE candidate failed",
			zap.String("roomID", roomID),
			zap.String("userID", userID),
			zap.String("peerID", peerID),
			zap.Error(err))
	}
}

func (handler *Handler) sendToUser(ctx context.Context, userID string, msg WebSocketMessage) {
	usr, err := handler.userService.GetUser(ctx, userID)
	if err != nil {
		handler.logger.Error("Get user failed", zap.String("userID", userID), zap.Error(err))
		return
	}

	if usr.Conn == nil {
		handler.logger.Warn("User connection is nil", zap.String("userID", userID))
		return
	}

	if err := usr.Conn.WriteJSON(msg); err != nil {
		handler.logger.Error("Write message to user failed",
			zap.String("userID", userID),
			zap.Error(err))
	}
}
