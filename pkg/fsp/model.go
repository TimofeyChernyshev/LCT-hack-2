package fsp

import (
	"time"

	"github.com/google/uuid"
)

// SportsRank represents official sports title or category in Russian Federation of Sports Programming
type SportsRank string

const (
	RankZMS         SportsRank = "ЗМС"                         // Заслуженный мастер спорта
	RankMSMK        SportsRank = "МСМК"                        // Мастер спорта международного класса
	RankMS          SportsRank = "Мастер спорта"               // Мастер спорта (МС)
	RankKMS         SportsRank = "КМС"                         // Кандидат в мастера спорта (КМС)
	Rank1           SportsRank = "1-й спортивный разряд"       // 1 разряд
	Rank2           SportsRank = "2-й спортивный разряд"       // 2 разряд
	Rank3           SportsRank = "3-й спортивный разряд"       // 3 разряд
	RankUnranked    SportsRank = "Без разряда"                 // Без разряда
)

// Discipline in sports programming
type Discipline string

const (
	DisciplineAlgorithms Discipline = "Алгоритмическое программирование"
	DisciplineProducts   Discipline = "Продуктовое программирование"
	DisciplineRobotics   Discipline = "Робототехника"
	DisciplineAI         Discipline = "Системы искусственного интеллекта"
)

// AchievementCategory represents tournament level/type
type AchievementCategory string

const (
	CategoryChampionship AchievementCategory = "championship" // Чемпионат России / мира
	CategoryCup          AchievementCategory = "cup"          // Кубок России / ФСП
	CategoryHackathon    AchievementCategory = "hackathon"    // Хакатоны ФСП
	CategoryRegional     AchievementCategory = "regional"     // Региональные первенства
	CategoryOlympiad     AchievementCategory = "olympiad"     // Олимпиады и студенческие лиги
)

// Achievement represents verified sporting accomplishment
type Achievement struct {
	ID          string              `json:"id"`
	ExternalID  string              `json:"externalId"`
	EventName   string              `json:"eventName"`
	EventDate   string              `json:"eventDate,omitempty"` // YYYY-MM-DD
	Place       *int                `json:"place,omitempty"`     // 1, 2, 3... nil for participant
	Category    AchievementCategory `json:"category"`
	Score       *float64            `json:"score,omitempty"`
	Weight      int                 `json:"weight"` // 1..10 ranking importance
	Badge       string              `json:"badge,omitempty"`  // gold, silver, bronze, winner, finalist, participant
	Description string              `json:"description,omitempty"`
	Payload     map[string]any      `json:"payload,omitempty"`
}

// Member represents athlete registered in Russian Sports Programming Federation registry
type Member struct {
	FSPID        string        `json:"fspId"`        // e.g. "FSP-RU-77-00101"
	FullName     string        `json:"fullName"`     // "Смирнов Александр Дмитриевич"
	SportsRank   SportsRank    `json:"sportsRank"`   // "Мастер спорта"
	Rating       int           `json:"rating"`       // ELO-like sports rating (e.g. 2480)
	Region       string        `json:"region"`       // "г. Москва"
	Discipline   Discipline    `json:"discipline"`   // "Алгоритмическое программирование"
	Status       string        `json:"status"`       // "active", "honorary"
	Verified     bool          `json:"verified"`
	Achievements []Achievement `json:"achievements"`
}

// VerificationRequest to verify participant credentials
type VerificationRequest struct {
	FSPID       string `json:"fspId"`
	FullName    string `json:"fullName,omitempty"`
	Certificate string `json:"certificate,omitempty"`
}

// VerificationResult returned by FSP registry verification
type VerificationResult struct {
	IsValid     bool        `json:"isValid"`
	Member      *Member     `json:"member,omitempty"`
	Message     string      `json:"message"`
	VerifiedAt  time.Time   `json:"verifiedAt"`
}

// ProfileState is the candidate profile enriched with FSP achievements
type ProfileState struct {
	UserID            uuid.UUID     `json:"userId"`
	HasFSP            bool          `json:"hasFsp"`
	FSPMemberID       *string       `json:"fspMemberId,omitempty"`
	FullName          *string       `json:"fullName,omitempty"`
	SportsRank        *string       `json:"sportsRank,omitempty"`
	FSPRating         int           `json:"fspRating"`
	Region            *string       `json:"region,omitempty"`
	FSPScore          float64       `json:"fspScore"` // S_fsp \in [0, 100]
	FSPWeightSum      int           `json:"fspWeightSum"`
	AchievementsCount int           `json:"achievementsCount"`
	BestPlace         *int          `json:"bestPlace,omitempty"`
	Achievements      []Achievement `json:"achievements"`
	VerificationSource string       `json:"verificationSource"` // "mock", "keycloak", "registry_api"
	LinkedAt          *time.Time    `json:"linkedAt,omitempty"`
	Explanation       string        `json:"explanation"`
}

// ScoreResult summarizes computed FSP influence for ranking and explainability
type ScoreResult struct {
	Score             float64 `json:"score"`             // S_fsp \in [0, 100]
	WeightSum         int     `json:"weightSum"`         // sum of achievement weights
	AchievementsCount int     `json:"achievementsCount"`
	BestPlace         *int    `json:"bestPlace,omitempty"`
	Explanation       string  `json:"explanation"`       // human-readable explainability text
}
