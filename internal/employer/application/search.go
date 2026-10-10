package application

import (
	"context"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/domain"
)

func (s *Service) SearchCandidates(ctx context.Context, authHeader string, q SearchQuery) (*SearchPage, error) {
	return s.search.SearchCandidates(ctx, authHeader, q)
}

func (s *Service) GetNeedMatches(ctx context.Context, ownerID, needID, authHeader string, q MatchesQuery) (*SearchPage, error) {
	// Проверяем, что need принадлежит компании владельца.
	c, err := s.companies.GetByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	n, err := s.needs.GetByID(ctx, needID)
	if err != nil {
		return nil, err
	}
	if n.CompanyID != c.ID {
		return nil, domain.ErrNotOwner
	}
	return s.search.MatchesForNeed(ctx, authHeader, needID, q)
}

func (s *Service) GetCandidateCard(ctx context.Context, authHeader, candidateUserID string) (*CandidateCard, error) {
	return s.candidate.GetCard(ctx, authHeader, candidateUserID)
}
