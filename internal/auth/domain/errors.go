package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrUserBlocked        = errors.New("user blocked")
	ErrUserLocked         = errors.New("user locked")
	ErrInvalidToken       = errors.New("invalid token")
	ErrConsentRequired    = errors.New("consent required")
	ErrInvalidRole        = errors.New("invalid role")
	ErrWeakPassword       = errors.New("weak password")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidConsentType = errors.New("invalid consent type")
)
