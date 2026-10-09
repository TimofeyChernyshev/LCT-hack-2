package application

import (
	"context"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
)

func (s *Service) GetMyFSP(ctx context.Context, userID string) (*domain.FSPState, error) {
	memberID, linkedAt, err := s.fsp.GetMemberID(ctx, userID)
	if err != nil {
		return nil, err
	}
	achievements, err := s.fsp.ListAchievements(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &domain.FSPState{
		FSPMemberID:  memberID,
		LinkedAt:     linkedAt,
		Achievements: achievements,
	}, nil
}

func (s *Service) LinkFSP(ctx context.Context, userID, memberID string) error {
	if memberID == "" {
		return domain.ErrInvalidInput
	}
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := s.fsp.LinkMemberID(ctx, userID, memberID); err != nil {
			return err
		}
		// В MVP создаём демо-достижение, чтобы UI было что показать.
		// В проде — вызов реестра ФСП.
		return s.fsp.CreateAchievement(ctx, &domain.FSPAchievement{
			UserID:    userID,
			EventName: "Mock FSP Event (linked)",
			Weight:    1,
			Source:    "mock",
		})
	})
}
