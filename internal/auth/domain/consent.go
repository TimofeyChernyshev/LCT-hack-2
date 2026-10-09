package domain

import "time"

type ConsentType string

const (
	ConsentPDNProcessing      ConsentType = "pdn_processing"
	ConsentProfilePublication ConsentType = "profile_publication"
)

func (c ConsentType) Valid() bool {
	switch c {
	case ConsentPDNProcessing, ConsentProfilePublication:
		return true
	}
	return false
}

type Consent struct {
	ID        string
	UserID    string
	Type      ConsentType
	Version   string
	GrantedAt time.Time
	RevokedAt *time.Time
	IP        string
	UserAgent string
}

func (c Consent) IsActive() bool { return c.RevokedAt == nil }
