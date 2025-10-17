package user

import (
	"context"
	"errors"
	"time"
	"vidcall/pkg/repository"

	"go.uber.org/fx"
)

type Service struct {
	repo repository.Repository[string, User]
}

type ServiceParams struct {
	fx.In

	Repository repository.Repository[string, User]
}

func NewService(params ServiceParams) *Service {
	return &Service{
		repo: params.Repository,
	}
}

func (service *Service) GetUser(ctx context.Context, ID string) (User, error) {
	return service.repo.Find(ctx, ID)
}

func (service *Service) CreateUser(ctx context.Context, user User) (User, error) {
	user.LastActive = time.Now()
	return service.repo.Insert(ctx, user)
}

func (service *Service) UpdateActive(ctx context.Context, userID string) (User, error) {
	user, err := service.repo.Find(ctx, userID)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return User{}, err
		}

		user = User{
			ID:         userID,
			LastActive: time.Now(),
		}
		return service.repo.Insert(ctx, user)
	}

	user.LastActive = time.Now()
	return service.repo.Update(ctx, user)
}

func (service *Service) DeleteUser(ctx context.Context, userID string) error {
	return service.repo.Delete(ctx, userID)
}

func (service *Service) ListUsers(ctx context.Context) ([]User, error) {
	return service.repo.FindList(ctx)
}

func (service *Service) UpdateUser(ctx context.Context, user User) (User, error) {
	user.LastActive = time.Now()
	return service.repo.Update(ctx, user)
}
