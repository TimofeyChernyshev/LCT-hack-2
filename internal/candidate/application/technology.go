package application

import (
	"context"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
)

type TechnologyInput struct {
	TechnologyID string
	Level        *int
}

func (s *Service) ReplaceTechnologies(ctx context.Context, userID string, in []TechnologyInput) error {
	techs := make([]domain.CandidateTechnology, 0, len(in))
	for _, t := range in {
		if t.Level != nil && (*t.Level < 1 || *t.Level > 5) {
			return domain.ErrInvalidTechnology
		}
		techs = append(techs, domain.CandidateTechnology{
			UserID:       userID,
			TechnologyID: t.TechnologyID,
			Level:        t.Level,
		})
	}
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		return s.technologies.ReplaceForUser(ctx, userID, techs)
	})
}
