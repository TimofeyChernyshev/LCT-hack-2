package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/mail"
	"strings"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/domain"
)

type RegisterInput struct {
	Email                string
	Password             string
	Role                 domain.Role
	ConsentPDN           bool
	ConsentProfilePublic bool
	IP                   string
	UserAgent            string
}

type RegisterOutput struct {
	UserID string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*RegisterOutput, error) {
	if _, err := mail.ParseAddress(in.Email); err != nil {
		return nil, domain.ErrInvalidEmail
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))

	if !in.Role.Valid() || !in.Role.CanSelfRegister() {
		return nil, domain.ErrInvalidRole
	}
	if len(in.Password) < s.cfg.MinPasswordLen {
		return nil, domain.ErrWeakPassword
	}
	if !in.ConsentPDN || !in.ConsentProfilePublic {
		return nil, domain.ErrConsentRequired
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	raw, hashTok, err := generateToken()
	if err != nil {
		return nil, err
	}
	exp := timeNow().Add(s.cfg.EmailVerificationTTL)

	var userID string
	err = s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		u := &domain.User{
			Email:        in.Email,
			PasswordHash: hash,
			Role:         in.Role,
			Status:       domain.StatusActive,
		}
		if err := s.users.Create(ctx, u); err != nil {
			return err
		}
		for _, c := range []struct {
			t   domain.ConsentType
			ver string
		}{
			{domain.ConsentPDNProcessing, s.cfg.ConsentVersionPDN},
			{domain.ConsentProfilePublication, s.cfg.ConsentVersionPub},
		} {
			if err := s.consents.Grant(ctx, &domain.Consent{
				UserID: u.ID, Type: c.t, Version: c.ver,
				IP: in.IP, UserAgent: in.UserAgent,
			}); err != nil {
				return fmt.Errorf("grant consent %s: %w", c.t, err)
			}
		}
		if err := s.verifTokens.Create(ctx, u.ID, hashTok, exp); err != nil {
			return fmt.Errorf("create verification token: %w", err)
		}
		userID = u.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Письмо отправляем после успешного коммита.
	link := s.cfg.AppBaseURL + "/auth/verify-email?token=" + raw
	if err := s.mailer.SendEmailVerification(ctx, in.Email, link); err != nil {
		s.logger.Error("send verification email", "err", err, "user_id", userID)
	}

	_ = s.audit.Log(ctx, userID, "auth.register", in.IP, in.UserAgent,
		map[string]any{"role": string(in.Role)})

	return &RegisterOutput{UserID: userID}, nil
}

func generateToken() (raw, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hashed = hex.EncodeToString(sum[:])
	return raw, hashed, nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
