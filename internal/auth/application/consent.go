package application

import (
	"context"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/domain"
)

func (s *Service) GrantConsent(ctx context.Context, userID string, t domain.ConsentType, ip, ua string) error {
	if !t.Valid() {
		return domain.ErrInvalidConsentType
	}
	ver := s.cfg.ConsentVersionPDN
	if t == domain.ConsentProfilePublication {
		ver = s.cfg.ConsentVersionPub
	}
	return s.consents.Grant(ctx, &domain.Consent{
		UserID: userID, Type: t, Version: ver, IP: ip, UserAgent: ua,
	})
}

func (s *Service) RevokeConsent(ctx context.Context, userID string, t domain.ConsentType) error {
	if !t.Valid() {
		return domain.ErrInvalidConsentType
	}
	return s.consents.Revoke(ctx, userID, t)
}

func (s *Service) ListConsents(ctx context.Context, userID string) ([]domain.Consent, error) {
	return s.consents.ListByUser(ctx, userID)
}
