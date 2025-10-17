package rest

import (
	"net/http"
	"time"

	"vidcall/internal/module/room"
	"vidcall/internal/module/view"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"
)

var Module = fx.Module("rest",
	fx.Provide(
		fx.Annotate(
			NewRouter,
			fx.As(new(http.Handler)),
		),
	),
)

type RouterParams struct {
	fx.In

	// Add other dependencies here if needed
	RoomHandler *room.Handler
	ViewHandler *view.Handler
}

func NewRouter(params RouterParams) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(middleware.Timeout(30 * time.Second))

	router.Get("/health", healthCheck)

	router.Get("/", params.ViewHandler.RenderHomepage)
	router.Get("/call", params.ViewHandler.RenderCallPage)

	router.Get("/rooms", params.RoomHandler.ListRooms)
	router.Post("/rooms", params.RoomHandler.CreateRoom)
	router.Get("/rooms/{roomID}", params.RoomHandler.GetRoom)
	router.Get("/rooms/{roomID}/ws", nil)
	router.Delete("/rooms/{roomID}", params.RoomHandler.DeleteRoom)

	return router
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
