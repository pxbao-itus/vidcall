package rtc

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"vidcall/internal/common"
	"vidcall/internal/module/room"
	"vidcall/internal/module/user"

	"github.com/gorilla/websocket"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("rtc",
	fx.Provide(NewHandler),
)

type Handler struct {
	roomService *room.Service
	userService *user.Service

	logger *zap.Logger
}

type HandlerParams struct {
	fx.In

	RoomService *room.Service
	UserService *user.Service
	Logger      *zap.Logger
}

func NewHandler(params HandlerParams) *Handler {
	return &Handler{
		roomService: params.RoomService,
		userService: params.UserService,
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
		HandshakeTimeout:  5 * time.Second,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
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

	if _, err := handler.userService.CreateUser(r.Context(), usr); err != nil {
		http.Error(w, "Create user failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := conn.Close(); err != nil {
			handler.logger.Error("WebSocket close failed", zap.String("userID", userID), zap.Error(err))
		}

		usr.Conn = nil
		if _, err := handler.userService.UpdateUser(r.Context(), usr); err != nil {
			handler.logger.Error("Update user conn to nil failed", zap.String("userID", userID), zap.Error(err))
		}
	}()

	commonRoom, err := handler.roomService.JoinRoom(r.Context(), roomID, userID)
	if err != nil {
		http.Error(w, "Join room failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	clientEvent := make(chan any, 1)
	go func() {
		for {
			select {
			case <-r.Context().Done():
				return
			default:
				var msg WebSocketMessage
				if err := usr.Conn.ReadJSON(&msg); err != nil {
					clientEvent <- fmt.Errorf("read msg: %w", err)
					return
				}

				handler.logger.Info("Received msg", zap.String("userID", userID), zap.Any("msg", msg))
				clientEvent <- msg
			}
		}
	}()

	var (
		peerUserMu sync.Mutex
		peerUser   *user.User
	)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, err := handler.userService.UpdateActive(r.Context(), userID); err != nil {
				handler.logger.Error("Update user active failed", zap.String("userID", userID), zap.Error(err))
				return
			}
		case <-r.Context().Done():
			if err := handler.roomService.LeaveRoom(r.Context(), commonRoom.ID, userID); err != nil {
				handler.logger.Error("Leave room failed", zap.String("roomID", commonRoom.ID), zap.String("userID", userID), zap.Error(err))
				return
			}

		case roomEvent := <-commonRoom.Subscribers[userID]:
			peerUserMu.Lock()
			if _, ok := roomEvent.LeaveRoom(); ok {
				peerUser = nil
			}

			if newComerID, ok := roomEvent.NewComer(); ok {
				peerUsr, err := handler.userService.GetUser(r.Context(), newComerID)
				if err != nil {
					handler.logger.Error("Get peer user failed", zap.String("roomID", commonRoom.ID), zap.String("userID", userID), zap.String("peerUserID", newComerID), zap.Error(err))
					return
				}
				peerUser = &peerUsr
			}

			if roomEvent.EventName == room.EventRoomDeleted {
				handler.logger.Info("Room deleted, closing connection", zap.String("roomID", commonRoom.ID), zap.String("userID", userID))
				return
			}
			peerUserMu.Unlock()
		case msg := <-clientEvent:
			if err, ok := msg.(error); ok {
				handler.logger.Error("Read msg failed", zap.String("roomID", commonRoom.ID), zap.String("userID", userID), zap.Error(err))
				return
			}

			clientMsg, ok := msg.(WebSocketMessage)
			if !ok {
				handler.logger.Error("Invalid msg type", zap.String("roomID", commonRoom.ID), zap.String("userID", userID))
				return
			}

			peerUserMu.Lock()
			if peerUser == nil {
				peerUserMu.Unlock()
				continue
			}

			if err := handler.handleClientMsg(peerUser.Conn, clientMsg); err != nil {
				handler.logger.Error("Handle client msg failed", zap.String("roomID", commonRoom.ID), zap.String("userID", userID), zap.Error(err))
				return
			}
			peerUserMu.Unlock()
		}

	}
}

func (handler *Handler) handleClientMsg(peerConn *websocket.Conn, msg WebSocketMessage) error {
	if peerConn == nil {
		return nil
	}

	switch msg.Event {
	case "offer", "answer", "candidate":
		return peerConn.WriteJSON(msg)
	case "hangup":
		// Notify the other client and clean up the room
	}

	return nil
}
