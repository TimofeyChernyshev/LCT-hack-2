package application

import (
	"context"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
)

func (s *Service) GetMyCategory(ctx context.Context, userID string) (*domain.CategoryState, error) {
	return s.category.GetState(ctx, userID)
}

func (s *Service) AssignCategory(ctx context.Context, userID, categoryID, gradeID, specializationID string) error {
	if categoryID == "" || gradeID == "" || specializationID == "" {
		return domain.ErrInvalidInput
	}
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		return s.category.Assign(ctx, userID, categoryID, gradeID, specializationID)
	})
}
