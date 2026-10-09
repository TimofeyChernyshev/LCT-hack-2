package application

import (
	"context"
	"strings"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/domain"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
)

type LoginInput struct {
	Email     string
	Password  string
	IP        string
	UserAgent string
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	UserID           string
	Role             domain.Role
	Email            string
	EmailVerified    bool
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*TokenPair, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	u, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if u.Status == domain.StatusBlocked {
		return nil, domain.ErrUserBlocked
	}
	if u.Status == domain.StatusDeleted {
		return nil, domain.ErrInvalidCredentials
	}
	now := time.Now()
	if u.IsLocked(now) {
		return nil, domain.ErrUserLocked
	}
	if !s.hasher.Verify(u.PasswordHash, in.Password) {
		_ = s.recordFailedLogin(ctx, u, now)
		_ = s.audit.Log(ctx, u.ID, "auth.login.failed", in.IP, in.UserAgent, nil)
		return nil, domain.ErrInvalidCredentials
	}
	if !u.IsEmailVerified() {
		return nil, domain.ErrEmailNotVerified
	}

	var pair *TokenPair
	err = s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := s.users.RecordSuccessfulLogin(ctx, u.ID, now); err != nil {
			return err
		}
		var err error
		pair, err = s.issueTokensInTx(ctx, u, in.IP, in.UserAgent)
		return err
	})
	if err != nil {
		return nil, err
	}
	_ = s.audit.Log(ctx, u.ID, "auth.login.success", in.IP, in.UserAgent, nil)
	return pair, nil
}

// recordFailedLogin — вся бизнес-логика про «когда блокировать» живёт здесь, не в SQL.
func (s *Service) recordFailedLogin(ctx context.Context, u *domain.User, now time.Time) error {
	failed := u.FailedLogins + 1
	var lockedUntil *time.Time
	if failed >= s.cfg.MaxFailedLogins {
		t := now.Add(s.cfg.LockDuration)
		lockedUntil = &t
	}
	return s.users.RecordFailedLogin(ctx, u.ID, failed, lockedUntil)
}

// issueTokensInTx используется внутри уже открытой транзакции.
// Здесь же — вспомогательный публичный враппер.
func (s *Service) issueTokensInTx(ctx context.Context, u *domain.User, ip, ua string) (*TokenPair, error) {
	now := time.Now()
	claims := jwtx.NewAccessClaims(
		u.ID, string(u.Role), u.Email, u.IsEmailVerified(),
		s.cfg.Issuer, s.cfg.AccessTTL,
	)
	access, err := s.signer.Sign(claims)
	if err != nil {
		return nil, err
	}
	rawRefresh, hashRefresh, err := generateToken()
	if err != nil {
		return nil, err
	}
	refreshExp := now.Add(s.cfg.RefreshTTL)
	if err := s.refresh.Create(ctx, u.ID, hashRefresh, refreshExp, ip, ua); err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:      access,
		RefreshToken:     rawRefresh,
		AccessExpiresAt:  now.Add(s.cfg.AccessTTL),
		RefreshExpiresAt: refreshExp,
		UserID:           u.ID,
		Role:             u.Role,
		Email:            u.Email,
		EmailVerified:    u.IsEmailVerified(),
	}, nil
}
