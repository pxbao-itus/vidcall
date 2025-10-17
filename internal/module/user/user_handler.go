package user

import (
	"net/http"

	"vidcall/internal/common"
	"vidcall/pkg/repository"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("user",
	fx.Provide(
		fx.Private,
		fx.Annotate(
			repository.NewSyncRepository[string, User],
			fx.As(new(repository.Repository[string, User])),
		),
	),
	fx.Provide(NewService),
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

func (handler *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := handler.service.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	if err := common.WriteResponse(w, http.StatusOK, users); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func (handler *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := common.GetParam(r, "userID")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	user, err := handler.service.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	if err := common.WriteResponse(w, http.StatusOK, user); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func (handler *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := common.BindRequest(r, &user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if _, err := handler.service.CreateUser(r.Context(), user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	if err := common.WriteResponse(w, http.StatusCreated, user); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
