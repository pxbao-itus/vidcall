package room

import (
	"context"

	"vidcall/pkg/repository"

	"go.uber.org/fx"
)

type Service struct {
	repo repository.Repository[string, Room]
}

type ServiceParams struct {
	fx.In

	Repository repository.Repository[string, Room]
}

func NewService(params ServiceParams) *Service {
	return &Service{
		repo: params.Repository,
	}
}

func (service *Service) CreateRoom(ctx context.Context, room Room) (Room, error) {
	return service.repo.Insert(ctx, room)
}

func (service *Service) GetRoom(ctx context.Context, id string) (Room, error) {
	return service.repo.Find(ctx, id)
}

func (service *Service) DeleteRoom(ctx context.Context, id string) error {
	return service.repo.Delete(ctx, id)
}

func (service *Service) ListRooms(ctx context.Context) ([]Room, error) {
	return service.repo.FindList(ctx)
}
