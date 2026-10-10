package application

import (
	"context"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
)

func (s *Service) GetMyVisibility(ctx context.Context, userID string) (*domain.Visibility, error) {
	v, err := s.visibility.Get(ctx, userID)
	if err != nil {
		d := domain.DefaultVisibility()
		return &d, nil
	}
	return v, nil
}

func (s *Service) UpdateMyVisibility(ctx context.Context, userID string, v *domain.Visibility) error {
	return s.visibility.Upsert(ctx, userID, v)
}
