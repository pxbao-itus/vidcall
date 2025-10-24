package room

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"vidcall/pkg/repository"

	"go.uber.org/fx"
)

var (
	ErrRoomIsFull    = errors.New("room is full")
	ErrRoomIsExpired = errors.New("room is expired")
	ErrUserNotInRoom = errors.New("user not in room")
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
	room.CreatedAt = time.Now().Unix()
	room.Subscribers = make(map[string]chan Event)
	return service.repo.Insert(ctx, room)
}

func (service *Service) GetRoom(ctx context.Context, id string) (Room, error) {
	return service.repo.Find(ctx, id)
}

func (service *Service) DeleteRoom(ctx context.Context, id string) error {
	room, err := service.GetRoom(ctx, id)
	if err != nil {
		return err
	}

	// Close the signal channel if it exists and have not been closed yet
	eventRoomDeleted := Event{EventName: EventRoomDeleted}
	room.emitEvent("", eventRoomDeleted)

	return service.repo.Delete(ctx, id)
}

func (service *Service) ListRooms(ctx context.Context) ([]Room, error) {
	rooms, err := service.repo.FindList(ctx)
	if err != nil {
		return nil, err
	}

	availableRooms := make([]Room, 0)
	for _, room := range rooms {
		if !room.IsExpired() {
			availableRooms = append(availableRooms, room)
			continue
		}

		if room.shouldDelete() {
			if err := service.DeleteRoom(ctx, room.ID); err != nil {
				return nil, fmt.Errorf("delete expired room %s: %w", room.ID, err)
			}
		}
	}

	// Sort rooms by CreatedAt ascending
	sort.Slice(availableRooms, func(i, j int) bool {
		return availableRooms[i].CreatedAt < availableRooms[j].CreatedAt
	})

	return availableRooms, nil
}

func (service *Service) ListOwnRooms(ctx context.Context, ownerID string) ([]Room, error) {
	rooms, err := service.ListRooms(ctx)
	if err != nil {
		return nil, err
	}

	if ownerID == "" {
		return rooms, nil
	}

	ownRooms := make([]Room, 0)
	for _, room := range rooms {
		if room.CreatedBy == ownerID {
			ownRooms = append(ownRooms, room)
		}
	}

	return ownRooms, nil
}

func (service *Service) JoinRoom(ctx context.Context, roomID, userID string) (Room, error) {
	room, err := service.repo.Find(ctx, roomID)
	if err != nil {
		return Room{}, err
	}

	if room.shouldDelete() {
		if err := service.DeleteRoom(ctx, roomID); err != nil {
			return Room{}, fmt.Errorf("delete expired room: %w", err)
		}

		return Room{}, ErrRoomIsExpired
	}

	if room.IsFull() {
		return Room{}, ErrRoomIsFull
	}

	room.AddGuest(userID)

	room, err = service.repo.Update(ctx, room)
	if err != nil {
		return Room{}, err
	}

	return room, nil
}

func (service *Service) LeaveRoom(ctx context.Context, roomID, userID string) error {
	room, err := service.repo.Find(ctx, roomID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		
		return err
	}

	room.RemoveGuest(userID)

	if room.shouldDelete() {
		if err := service.DeleteRoom(ctx, roomID); err != nil {
			return fmt.Errorf("delete expired room: %w", err)
		}

		return nil
	}

	if _, err := service.repo.Update(ctx, room); err != nil {
		return err
	}

	return nil
}

func (service *Service) EmitEvent(roomID, userID string, event Event) {
	room, err := service.repo.Find(context.Background(), roomID)
	if err != nil {
		return
	}

	room.emitEvent(userID, event)
}
