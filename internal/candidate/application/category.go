package application

import (
	"context"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
)

func (s *Service) GetMyCategory(ctx context.Context, userID string) (*domain.CategoryState, error) {
	return s.category.GetState(ctx, userID)
}
