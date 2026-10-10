package domain

import "time"

type VerificationToken struct {
	ID        string
	UserID    string
	Consumed  bool
	ExpiresAt time.Time
}

type PasswordResetToken struct {
	ID        string
	UserID    string
	Consumed  bool
	ExpiresAt time.Time
}

type RefreshToken struct {
	ID        string
	UserID    string
	Revoked   bool
	ExpiresAt time.Time
}
