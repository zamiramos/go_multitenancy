package car

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"sync"
)

type Service interface {
	PostCar(ctx context.Context, car Car) error
	GetCar(ctx context.Context, id string) (Car, error)
	DeleteCar(ctx context.Context, id string) error
}

// Car represents a single car.
// ID should be globally unique.
type Car struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type Repository interface {
	CreateCar(ctx context.Context, car Car) error
	GetCar(ctx context.Context, id string) (Car, error)
	DeleteCar(ctx context.Context, id string) error
}

var (
	ErrInconsistentIDs = errors.New("inconsistent IDs")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
)

type inmemService struct {
	mtx  sync.RWMutex
	repo Repository
}

func NewInmemService(repo Repository) Service {
	return &inmemService{
		repo: repo,
	}
}

func (s *inmemService) PostCar(ctx context.Context, car Car) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, err := s.repo.GetCar(ctx, car.ID); err != gorm.ErrRecordNotFound {
		return ErrAlreadyExists // POST = create, don't overwrite
	}
	err := s.repo.CreateCar(ctx, car)
	return err
}

func (s *inmemService) GetCar(ctx context.Context, id string) (Car, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	car, err := s.repo.GetCar(ctx, id)

	if err == gorm.ErrRecordNotFound {
		return car, ErrNotFound
	}

	return car, err
}

func (s *inmemService) DeleteCar(ctx context.Context, id string) error {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	err := s.repo.DeleteCar(ctx, id)

	if err != nil {
		return ErrNotFound
	}

	return nil
}
