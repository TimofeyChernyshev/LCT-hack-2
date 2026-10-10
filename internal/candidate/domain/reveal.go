package domain

import "time"

type RevealEntityType string

const (
	RevealEntityInvitation  RevealEntityType = "invitation"
	RevealEntityApplication RevealEntityType = "application"
)

type RevealReason string

const (
	RevealReasonInvitationAccepted RevealReason = "invitation_accepted"
	RevealReasonApplicationSent    RevealReason = "application_sent"
)

type ContactReveal struct {
	ID              string
	CandidateUserID string
	EmployerUserID  string
	EntityType      RevealEntityType
	EntityID        string
	Reason          RevealReason
	RevealedAt      time.Time
}
