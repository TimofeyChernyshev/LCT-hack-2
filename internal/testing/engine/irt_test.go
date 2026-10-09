package engine

import (
	"testing"
	"time"

	"TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
	"github.com/google/uuid"
)

func TestIRTScorer_ComputeSessionResult(t *testing.T) {
	scorer := NewIRTScorer(0.7, 0.9)

	// 1. High score -> upgrade_offered
	itemsHigh := []ItemEvaluation{
		{Discrimination: 1.2, Difficulty: -1.0, Score: 1.0},
		{Discrimination: 1.5, Difficulty: 0.0, Score: 1.0},
		{Discrimination: 1.8, Difficulty: 1.0, Score: 1.0},
		{Discrimination: 1.5, Difficulty: 1.5, Score: 1.0},
	}
	scoreHigh, thetaHigh, decisionHigh := scorer.ComputeSessionResult(itemsHigh)
	if scoreHigh != 100.0 {
		t.Errorf("expected 100.0 score, got %v", scoreHigh)
	}
	if decisionHigh != "upgrade_offered" {
		t.Errorf("expected upgrade_offered, got %s", decisionHigh)
	}
	if thetaHigh <= 0.0 {
		t.Errorf("expected positive theta for all correct answers, got %v", thetaHigh)
	}

	// 2. Medium score (75%) -> confirmed
	itemsMed := []ItemEvaluation{
		{Discrimination: 1.0, Difficulty: -1.0, Score: 1.0},
		{Discrimination: 1.0, Difficulty: 0.0, Score: 1.0},
		{Discrimination: 1.0, Difficulty: 0.5, Score: 1.0},
		{Discrimination: 1.0, Difficulty: 1.5, Score: 0.0},
	}
	scoreMed, _, decisionMed := scorer.ComputeSessionResult(itemsMed)
	if scoreMed != 75.0 {
		t.Errorf("expected 75.0 score, got %v", scoreMed)
	}
	if decisionMed != "confirmed" {
		t.Errorf("expected confirmed, got %s", decisionMed)
	}

	// 3. Low score (25%) -> downgrade_offered
	itemsLow := []ItemEvaluation{
		{Discrimination: 1.0, Difficulty: -1.0, Score: 1.0},
		{Discrimination: 1.0, Difficulty: 0.0, Score: 0.0},
		{Discrimination: 1.0, Difficulty: 0.5, Score: 0.0},
		{Discrimination: 1.0, Difficulty: 1.5, Score: 0.0},
	}
	scoreLow, thetaLow, decisionLow := scorer.ComputeSessionResult(itemsLow)
	if scoreLow != 25.0 {
		t.Errorf("expected 25.0 score, got %v", scoreLow)
	}
	if decisionLow != "downgrade_offered" {
		t.Errorf("expected downgrade_offered, got %s", decisionLow)
	}
	if thetaLow >= thetaHigh {
		t.Errorf("expected low theta %v < high theta %v", thetaLow, thetaHigh)
	}
}

func TestAntiCheatEngine_ValidateSessionTime(t *testing.T) {
	ac := NewAntiCheatEngine(AntiCheatConfig{
		SessionDuration:      10 * time.Minute,
		MinAnswerTimeSeconds: 2,
	})

	// Active session
	activeSession := &domain.Session{
		ID:        uuid.New(),
		StartedAt: time.Now().Add(-5 * time.Minute),
	}
	if err := ac.ValidateSessionTime(activeSession); err != nil {
		t.Errorf("expected active session to be valid, got: %v", err)
	}

	// Expired session
	expiredSession := &domain.Session{
		ID:        uuid.New(),
		StartedAt: time.Now().Add(-15 * time.Minute),
	}
	if err := ac.ValidateSessionTime(expiredSession); err != ErrSessionExpired {
		t.Errorf("expected ErrSessionExpired, got: %v", err)
	}
}

func TestAntiCheatEngine_CheckAnswerTiming(t *testing.T) {
	ac := NewAntiCheatEngine(AntiCheatConfig{
		SessionDuration:      30 * time.Minute,
		MinAnswerTimeSeconds: 2,
	})

	// Code answered in 500ms -> suspicious
	if ac.CheckAnswerTiming(domain.TaskTypeCode, 500*time.Millisecond) {
		t.Errorf("expected code answered in 500ms to be flagged suspicious")
	}

	// Code answered in 10s -> ok
	if !ac.CheckAnswerTiming(domain.TaskTypeCode, 10*time.Second) {
		t.Errorf("expected code answered in 10s to be approved")
	}
}
