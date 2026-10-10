package application

import (
	"log/slog"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
)

type Config struct {
	Issuer               string
	AccessTTL            time.Duration
	RefreshTTL           time.Duration
	EmailVerificationTTL time.Duration
	PasswordResetTTL     time.Duration
	AppBaseURL           string
	MaxFailedLogins      int
	LockDuration         time.Duration
	ConsentVersionPDN    string
	ConsentVersionPub    string
	MinPasswordLen       int
}

type Service struct {
	users       UserRepository
	verifTokens VerificationTokenRepository
	resetTokens PasswordResetRepository
	refresh     RefreshTokenRepository
	consents    ConsentRepository
	audit       AuditRepository
	txm         TransactionManager
	mailer      Mailer
	hasher      PasswordHasher
	signer      jwtx.Signer
	logger      *slog.Logger
	cfg         Config
}

func NewService(
	users UserRepository,
	verifTokens VerificationTokenRepository,
	resetTokens PasswordResetRepository,
	refresh RefreshTokenRepository,
	consents ConsentRepository,
	audit AuditRepository,
	txm TransactionManager,
	mailer Mailer,
	hasher PasswordHasher,
	signer jwtx.Signer,
	cfg Config,
) *Service {
	if cfg.MinPasswordLen == 0 {
		cfg.MinPasswordLen = 8
	}
	if cfg.MaxFailedLogins == 0 {
		cfg.MaxFailedLogins = 5
	}
	if cfg.LockDuration == 0 {
		cfg.LockDuration = 15 * time.Minute
	}
	return &Service{
		users: users, verifTokens: verifTokens, resetTokens: resetTokens,
		refresh: refresh, consents: consents, audit: audit, txm: txm,
		mailer: mailer, hasher: hasher, signer: signer, cfg: cfg,
	}
}
