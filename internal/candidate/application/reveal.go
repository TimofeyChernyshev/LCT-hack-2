package application

import (
	"context"
	"errors"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

type RecordRevealInput struct {
	CandidateUserID string
	EmployerUserID  string
	EntityType      domain.RevealEntityType
	EntityID        string
	Reason          domain.RevealReason
}

func (s *Service) RecordReveal(ctx context.Context, in RecordRevealInput) error {
	err := s.reveals.Create(ctx, &domain.ContactReveal{
		CandidateUserID: in.CandidateUserID,
		EmployerUserID:  in.EmployerUserID,
		EntityType:      in.EntityType,
		EntityID:        in.EntityID,
		Reason:          in.Reason,
	})
	// идемпотентно: если запись уже есть - ок
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil
	}
	return err
}
