package room

import (
	"net/http"

	"vidcall/internal/common"
	"vidcall/pkg/repository"

	"github.com/oklog/ulid/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("room",
	fx.Provide(
		fx.Private,
		fx.Annotate(
			repository.NewSyncRepository[string, Room],
			fx.As(new(repository.Repository[string, Room])),
		),
	),
	fx.Provide(
		fx.Private,
		NewService,
	),
	fx.Provide(NewHandler),
)

type Handler struct {
	service *Service
	logger  *zap.Logger
}

type HandlerParams struct {
	fx.In

	Logger  *zap.Logger
	Service *Service
}

func NewHandler(params HandlerParams) *Handler {
	return &Handler{
		service: params.Service,
		logger:  params.Logger,
	}
}

func (handler *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var room Room
	if err := common.BindRequest(r, &room); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	room.ID = ulid.Make().String()
	if _, err := handler.service.CreateRoom(r.Context(), room); err != nil {
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	if err := common.WriteResponse(w, http.StatusCreated, room); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func (handler *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomID")
	if roomID == "" {
		http.Error(w, "Missing room ID", http.StatusBadRequest)
		return
	}

	room, err := handler.service.GetRoom(r.Context(), roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	if err := common.WriteResponse(w, http.StatusOK, room); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func (handler *Handler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomID")
	if roomID == "" {
		http.Error(w, "Missing room ID", http.StatusBadRequest)
		return
	}

	if err := handler.service.DeleteRoom(r.Context(), roomID); err != nil {
		http.Error(w, "Failed to delete room", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := handler.service.ListRooms(r.Context())
	if err != nil {
		http.Error(w, "Failed to list rooms", http.StatusInternalServerError)
		return
	}

	if err := common.WriteResponse(w, http.StatusOK, rooms); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
