package domain

import "time"

type Role string

const (
	RoleCandidate Role = "candidate"
	RoleEmployer  Role = "employer"
	RoleAdmin     Role = "admin"
)

func (r Role) Valid() bool {
	switch r {
	case RoleCandidate, RoleEmployer, RoleAdmin:
		return true
	}
	return false
}

func (r Role) CanSelfRegister() bool {
	return r == RoleCandidate || r == RoleEmployer
}

type Status string

const (
	StatusActive  Status = "active"
	StatusBlocked Status = "blocked"
	StatusDeleted Status = "deleted"
)

type User struct {
	ID              string
	Email           string
	EmailVerifiedAt *time.Time
	PasswordHash    string
	Role            Role
	Status          Status
	FailedLogins    int
	LockedUntil     *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (u User) IsEmailVerified() bool { return u.EmailVerifiedAt != nil }

func (u User) IsLocked(now time.Time) bool {
	return u.LockedUntil != nil && u.LockedUntil.After(now)
}

func (u User) CanLogin() bool {
	return u.Status == StatusActive
}
