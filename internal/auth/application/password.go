package application

import (
	"context"
	"fmt"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/domain"
)

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}
	raw, hash, err := generateToken()
	if err != nil {
		return err
	}
	exp := time.Now().Add(s.cfg.PasswordResetTTL)
	if err := s.resetTokens.Create(ctx, u.ID, hash, exp); err != nil {
		return err
	}
	link := fmt.Sprintf("%s/auth/reset-password?token=%s", s.cfg.AppBaseURL, raw)
	_ = s.mailer.SendPasswordReset(ctx, u.Email, link)
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < s.cfg.MinPasswordLen {
		return domain.ErrWeakPassword
	}
	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	now := time.Now()
	tokenHash := hashToken(rawToken)

	var userID string
	err = s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		tok, err := s.resetTokens.GetForUpdate(ctx, tokenHash)
		if err != nil {
			return err
		}
		if tok.Consumed || now.After(tok.ExpiresAt) {
			return domain.ErrInvalidToken
		}
		if err := s.resetTokens.MarkConsumed(ctx, tok.ID); err != nil {
			return err
		}
		if err := s.users.UpdatePassword(ctx, tok.UserID, hash); err != nil {
			return err
		}
		if err := s.refresh.RevokeAllForUser(ctx, tok.UserID); err != nil {
			return err
		}
		userID = tok.UserID
		return nil
	})
	if err != nil {
		return err
	}
	_ = s.audit.Log(ctx, userID, "auth.password.reset", "", "", nil)
	return nil
}
