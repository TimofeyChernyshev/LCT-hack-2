package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/fsp"
)

// CandidateFSPProfile represents candidate's sports programming profile in testing service
type CandidateFSPProfile struct {
	UserID             uuid.UUID         `json:"userId"`
	FSPMemberID        *string           `json:"fspMemberId,omitempty"`
	FullName           *string           `json:"fullName,omitempty"`
	SportsRank         *string           `json:"sportsRank,omitempty"`
	FSPRating          int               `json:"fspRating"`
	Region             *string           `json:"region,omitempty"`
	Discipline         *string           `json:"discipline,omitempty"`
	HasFSP             bool              `json:"hasFsp"`
	FSPScore           float64           `json:"fspScore"`
	FSPWeightSum       int               `json:"fspWeightSum"`
	AchievementsCount  int               `json:"achievementsCount"`
	BestPlace          *int              `json:"bestPlace,omitempty"`
	VerificationSource string            `json:"verificationSource"`
	LinkedAt           *time.Time        `json:"linkedAt,omitempty"`
	Explanation        string            `json:"explanation"`
	CreatedAt          time.Time         `json:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt"`
	Achievements       []fsp.Achievement `json:"achievements"`
}

// LinkFSPInput request payload to link account with FSP registry ID
type LinkFSPInput struct {
	FSPMemberID string `json:"fspMemberId"`
}
