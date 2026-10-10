package engine

import (
	"errors"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
)

var (
	ErrSessionExpired = errors.New("test session has expired")
	ErrSuspiciousFast = errors.New("answer submitted suspiciously fast")
)

type AntiCheatConfig struct {
	SessionDuration      time.Duration
	MinAnswerTimeSeconds int // Minimum seconds expected per question
}

type AntiCheatEngine struct {
	config AntiCheatConfig
}

func NewAntiCheatEngine(cfg AntiCheatConfig) *AntiCheatEngine {
	if cfg.SessionDuration == 0 {
		cfg.SessionDuration = 45 * time.Minute
	}
	if cfg.MinAnswerTimeSeconds == 0 {
		cfg.MinAnswerTimeSeconds = 2
	}
	return &AntiCheatEngine{config: cfg}
}

// ValidateSessionTime checks if the session is still active and within the time limit
func (ac *AntiCheatEngine) ValidateSessionTime(session *domain.Session) error {
	deadline := session.StartedAt.Add(ac.config.SessionDuration)
	if time.Now().After(deadline) {
		return ErrSessionExpired
	}
	return nil
}

// CheckAnswerTiming detects possible automated script or speedrunning anomalies
func (ac *AntiCheatEngine) CheckAnswerTiming(taskType domain.TaskType, elapsed time.Duration) bool {
	// For complex questions like code or sql, answering under 2 seconds is suspicious
	if (taskType == domain.TaskTypeCode || taskType == domain.TaskTypeSQL) && elapsed < time.Duration(ac.config.MinAnswerTimeSeconds)*time.Second {
		return false
	}
	return true
}
