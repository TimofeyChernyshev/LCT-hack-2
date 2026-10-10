package application

import (
	"context"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
)

type AddExperienceInput struct {
	Company     string
	Position    string
	StartedAt   time.Time
	EndedAt     time.Time
	Description *string
}

func (s *Service) ListExperiences(ctx context.Context, userID string) ([]domain.Experience, error) {
	return s.experiences.ListByUser(ctx, userID)
}

func (s *Service) DeleteExperience(ctx context.Context, userID, id string) error {
	return s.experiences.Delete(ctx, userID, id)
}

func (s *Service) AddExperience(ctx context.Context, userID string, in AddExperienceInput) (*domain.Experience, error) {
	e := &domain.Experience{
		UserID:      userID,
		Company:     in.Company,
		Position:    in.Position,
		StartedAt:   in.StartedAt,
		EndedAt:     in.EndedAt,
		Description: in.Description,
	}
	if err := s.experiences.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}
