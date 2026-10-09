package application

import (
	"context"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/domain"
)

type Me struct {
	ID              string
	Email           string
	Role            domain.Role
	Status          domain.Status
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
}

func (s *Service) GetMe(ctx context.Context, userID string) (*Me, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &Me{
		ID: u.ID, Email: u.Email, Role: u.Role, Status: u.Status,
		EmailVerifiedAt: u.EmailVerifiedAt, CreatedAt: u.CreatedAt,
	}, nil
}

func (s *Service) DeleteMe(ctx context.Context, userID, password, ip, ua string) error {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !s.hasher.Verify(u.PasswordHash, password) {
		return domain.ErrInvalidCredentials
	}
	err = s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := s.users.SoftDelete(ctx, userID); err != nil {
			return err
		}
		return s.refresh.RevokeAllForUser(ctx, userID)
	})
	if err != nil {
		return err
	}
	_ = s.audit.Log(ctx, userID, "auth.account.deleted", ip, ua, nil)
	return nil
}
