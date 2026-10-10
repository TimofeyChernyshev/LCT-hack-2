package application

import (
	"context"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/domain"
)

func (s *Service) Refresh(ctx context.Context, rawRefresh string, ip, ua string) (*TokenPair, error) {
	now := time.Now()
	hash := hashToken(rawRefresh)

	var pair *TokenPair
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		tok, err := s.refresh.GetForUpdate(ctx, hash)
		if err != nil {
			return err
		}
		if tok.Revoked || now.After(tok.ExpiresAt) {
			return domain.ErrInvalidToken
		}
		if err := s.refresh.MarkRevoked(ctx, tok.ID); err != nil {
			return err
		}
		u, err := s.users.GetByID(ctx, tok.UserID)
		if err != nil {
			return err
		}
		if u.Status != domain.StatusActive {
			return domain.ErrUserBlocked
		}
		pair, err = s.issueTokensInTx(ctx, u, ip, ua)
		return err
	})
	return pair, err
}

func (s *Service) Logout(ctx context.Context, rawRefresh string) error {
	hash := hashToken(rawRefresh)
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		tok, err := s.refresh.GetForUpdate(ctx, hash)
		if err != nil {
			return nil // logout идемпотентен
		}
		if tok.Revoked {
			return nil
		}
		return s.refresh.MarkRevoked(ctx, tok.ID)
	})
}

func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	return s.refresh.RevokeAllForUser(ctx, userID)
}
