package view

//
//import (
//	"context"
//	"html/template"
//	"net/http"
//	"os"
//
//	"go.uber.org/fx"
//	"go.uber.org/zap"
//)
//
//var Module = fx.Module("view",
//	fx.Provide(NewHandler),
//)
//
//type Handler struct {
//	template    *template.Template
//	logger      *zap.Logger
//	roomService RoomService
//}
//
//type RoomService interface {
//	ListRooms(ctx context.Context) ([]interface{}, error)
//}
//
//type HandlerParams struct {
//	fx.In
//
//	Logger      *zap.Logger
//	RoomService RoomService
//}
//
//func NewHandler(params HandlerParams) *Handler {
//	wd, err := os.Getwd()
//	if err != nil {
//		params.Logger.Fatal("Failed to get working directory", zap.Error(err))
//	}
//
//		template:    tmpl,
//		logger:      params.Logger,
//		roomService: params.RoomService,
//	if err != nil {
//		params.Logger.Fatal("Failed to parse templates", zap.Error(err))
//	}
//
//	return &Handler{
//
//	rooms, err := handler.roomService.ListRooms(r.Context())
//	if err != nil {
//		handler.logger.Error("Failed to list rooms", zap.Error(err))
//		rooms = []interface{}{}
//	}
//
//	data := map[string]interface{}{
//		"Rooms": rooms,
//	}
//
//	if err := handler.template.ExecuteTemplate(w, "index.html", data); err != nil {
//	}
//}
//
//func (handler *Handler) RenderHomepage(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "text/html; charset=utf-8")
//	if err := handler.template.ExecuteTemplate(w, "index.html", nil); err != nil {
//		http.Error(w, "Failed to render template", http.StatusInternalServerError)
//		return
//	}
//}
//
//func (handler *Handler) RenderCallPage(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "text/html; charset=utf-8")
//	if err := handler.template.ExecuteTemplate(w, "call.html", nil); err != nil {
//		http.Error(w, "Failed to render template", http.StatusInternalServerError)
//		return
//	}
//}
