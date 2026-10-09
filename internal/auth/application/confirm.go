package application

import (
	"context"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/domain"
)

func (s *Service) ConfirmEmail(ctx context.Context, rawToken string) error {
	now := time.Now()
	hash := hashToken(rawToken)

	var userID string
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		tok, err := s.verifTokens.GetForUpdate(ctx, hash)
		if err != nil {
			return err
		}
		if tok.Consumed || now.After(tok.ExpiresAt) {
			return domain.ErrInvalidToken
		}
		if err := s.verifTokens.MarkConsumed(ctx, tok.ID); err != nil {
			return err
		}
		if err := s.users.MarkEmailVerified(ctx, tok.UserID, now); err != nil {
			return err
		}
		userID = tok.UserID
		return nil
	})
	if err != nil {
		return err
	}
	_ = s.audit.Log(ctx, userID, "auth.email.verified", "", "", nil)
	return nil
}

func (s *Service) ResendConfirmation(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}
	if u.IsEmailVerified() {
		return nil
	}
	_ = s.verifTokens.InvalidateAll(ctx, u.ID)

	raw, hashTok, err := generateToken()
	if err != nil {
		return err
	}
	exp := time.Now().Add(s.cfg.EmailVerificationTTL)
	if err := s.verifTokens.Create(ctx, u.ID, hashTok, exp); err != nil {
		return err
	}
	link := s.cfg.AppBaseURL + "/auth/verify-email?token=" + raw
	_ = s.mailer.SendEmailVerification(ctx, u.Email, link)
	return nil
}
