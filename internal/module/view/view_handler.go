package view

import (
	"html/template"
	"net/http"
	"os"
	"time"

	"vidcall/internal/module/room"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("view",
	fx.Provide(NewHandler),
)

type Handler struct {
	template    *template.Template
	logger      *zap.Logger
	roomService *room.Service
}

type HandlerParams struct {
	fx.In

	Logger      *zap.Logger
	RoomService *room.Service
}

func NewHandler(params HandlerParams) *Handler {
	wd, err := os.Getwd()
	if err != nil {
		params.Logger.Fatal("Failed to get working directory", zap.Error(err))
	}

	tmpl, err := template.ParseGlob(wd + "/internal/template/*.html")
	if err != nil {
		params.Logger.Fatal("Failed to parse templates", zap.Error(err))
	}

	return &Handler{
		template:    tmpl,
		logger:      params.Logger,
		roomService: params.RoomService,
	}
}

func (handler *Handler) RenderHomepage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	rooms, err := handler.roomService.ListRooms(r.Context())
	if err != nil {
		handler.logger.Error("Failed to list rooms", zap.Error(err))
		rooms = []room.Room{}
	}

	// Generate a random ID for new users
	randomID := time.Now().Unix()

	data := map[string]interface{}{
		"Rooms":    rooms,
		"RandomID": randomID,
	}

	if err := handler.template.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func (handler *Handler) RenderCallPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get roomID from URL path using chi URLParam
	roomID := r.PathValue("roomID")

	data := map[string]interface{}{
		"RoomID": roomID,
	}

	if err := handler.template.ExecuteTemplate(w, "call.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
