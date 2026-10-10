package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/infrastructure/config"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/fsp"
)

func TestFSPService_LinkFSP_Success(t *testing.T) {
	repo := newMockRepository()
	cfg := &config.Config{
		GradeChangeCooldownDays: 90,
		ItemsPerSession:         10,
		PassThresholdRatio:      0.7,
		UpgradeThresholdRatio:   0.9,
	}
	svc := NewTestingService(repo, cfg)
	ctx := context.Background()

	userID := uuid.New()

	// Link with canonical FSP ID: Master of Sports Alexander Smirnov
	profile, err := svc.LinkFSP(ctx, userID, "FSP-RU-77-00101", "manual")
	require.NoError(t, err)
	require.NotNil(t, profile)

	assert.Equal(t, userID, profile.UserID)
	assert.True(t, profile.HasFSP)
	require.NotNil(t, profile.FSPMemberID)
	assert.Equal(t, "FSP-RU-77-00101", *profile.FSPMemberID)
	require.NotNil(t, profile.FullName)
	assert.Equal(t, "Смирнов Александр Дмитриевич", *profile.FullName)
	require.NotNil(t, profile.SportsRank)
	assert.Equal(t, "Мастер спорта", *profile.SportsRank)
	assert.Equal(t, 2540, profile.FSPRating)
	assert.Greater(t, profile.FSPScore, 80.0)
	assert.Equal(t, 25, profile.FSPWeightSum)
	assert.Equal(t, 3, profile.AchievementsCount)
	require.NotNil(t, profile.BestPlace)
	assert.Equal(t, 1, *profile.BestPlace)
	assert.Contains(t, profile.Explanation, "Мастер спорта")
	assert.Contains(t, profile.Explanation, "2540")
	assert.Len(t, profile.Achievements, 3)

	// Fetch profile again via GetCandidateFSP
	fetched, err := svc.GetCandidateFSP(ctx, userID)
	require.NoError(t, err)
	assert.True(t, fetched.HasFSP)
	assert.Equal(t, profile.FSPScore, fetched.FSPScore)
	assert.Equal(t, profile.Explanation, fetched.Explanation)
}

func TestFSPService_LinkFSP_Alias(t *testing.T) {
	repo := newMockRepository()
	cfg := &config.Config{
		PassThresholdRatio:    0.7,
		UpgradeThresholdRatio: 0.9,
	}
	svc := NewTestingService(repo, cfg)
	ctx := context.Background()

	userID := uuid.New()

	// Link using short alias FSP-10002 -> resolves to Maria Ivanova (KMS)
	profile, err := svc.LinkFSP(ctx, userID, "FSP-10002", "manual")
	require.NoError(t, err)
	require.NotNil(t, profile)

	assert.True(t, profile.HasFSP)
	assert.Equal(t, "FSP-RU-78-00202", *profile.FSPMemberID)
	assert.Equal(t, "Иванова Мария Сергеевна", *profile.FullName)
	assert.Equal(t, "КМС", *profile.SportsRank)
	assert.Equal(t, 2190, profile.FSPRating)
	assert.Equal(t, 1, *profile.BestPlace)
}

func TestFSPService_LinkFSP_NotFound(t *testing.T) {
	repo := newMockRepository()
	cfg := &config.Config{
		PassThresholdRatio:    0.7,
		UpgradeThresholdRatio: 0.9,
	}
	svc := NewTestingService(repo, cfg)
	ctx := context.Background()

	userID := uuid.New()

	_, err := svc.LinkFSP(ctx, userID, "UNKNOWN-FSP-9999", "manual")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrFSPMemberNotFound)
}

func TestFSPService_NoFSPEdgeCase(t *testing.T) {
	repo := newMockRepository()
	cfg := &config.Config{
		PassThresholdRatio:    0.7,
		UpgradeThresholdRatio: 0.9,
	}
	svc := NewTestingService(repo, cfg)
	ctx := context.Background()

	// Candidate who has never linked FSP
	freshUserID := uuid.New()

	profile, err := svc.GetCandidateFSP(ctx, freshUserID)
	// ТЗ: Обработка краевого случая: корректная работа профиля кандидата без истории ФСП
	// (отсутствие ошибок, нейтральное состояние, базовый скоринг).
	require.NoError(t, err)
	require.NotNil(t, profile)

	assert.Equal(t, freshUserID, profile.UserID)
	assert.False(t, profile.HasFSP)
	assert.Nil(t, profile.FSPMemberID)
	assert.Nil(t, profile.FullName)
	assert.Nil(t, profile.SportsRank)
	assert.Equal(t, 0, profile.FSPRating)
	assert.Equal(t, 0.0, profile.FSPScore)
	assert.Equal(t, 0, profile.FSPWeightSum)
	assert.Equal(t, 0, profile.AchievementsCount)
	assert.Nil(t, profile.BestPlace)
	assert.Empty(t, profile.Achievements)
	assert.Equal(t, fsp.NoFSPExplanation, profile.Explanation)
}

func TestFSPService_UnlinkFSP(t *testing.T) {
	repo := newMockRepository()
	cfg := &config.Config{
		PassThresholdRatio:    0.7,
		UpgradeThresholdRatio: 0.9,
	}
	svc := NewTestingService(repo, cfg)
	ctx := context.Background()

	userID := uuid.New()

	// 1. Link FSP
	_, err := svc.LinkFSP(ctx, userID, "FSP-RU-16-00303", "manual")
	require.NoError(t, err)

	linked, err := svc.GetCandidateFSP(ctx, userID)
	require.NoError(t, err)
	assert.True(t, linked.HasFSP)

	// 2. Unlink FSP
	unlinked, err := svc.UnlinkFSP(ctx, userID)
	require.NoError(t, err)
	assert.False(t, unlinked.HasFSP)
	assert.Nil(t, unlinked.FSPMemberID)
	assert.Equal(t, 0.0, unlinked.FSPScore)
	assert.Equal(t, fsp.NoFSPExplanation, unlinked.Explanation)

	// 3. Confirm via GetCandidateFSP
	check, err := svc.GetCandidateFSP(ctx, userID)
	require.NoError(t, err)
	assert.False(t, check.HasFSP)
	assert.Equal(t, 0.0, check.FSPScore)
}

func TestFSPService_SyncKeycloakFSP(t *testing.T) {
	repo := newMockRepository()
	cfg := &config.Config{
		PassThresholdRatio:    0.7,
		UpgradeThresholdRatio: 0.9,
	}
	svc := NewTestingService(repo, cfg)
	ctx := context.Background()

	userID := uuid.New()

	// Keycloak JWT claims with fsp_id
	claims := map[string]any{
		"iss":    "http://keycloak.fsp.local/realms/fsp",
		"sub":    userID.String(),
		"fsp_id": "FSP-RU-54-00404",
	}

	profile, err := svc.SyncKeycloakFSP(ctx, userID, claims)
	require.NoError(t, err)
	require.NotNil(t, profile)

	assert.True(t, profile.HasFSP)
	assert.Equal(t, "FSP-RU-54-00404", *profile.FSPMemberID)
	assert.Equal(t, "keycloak", profile.VerificationSource)
	assert.Equal(t, "Ковалев Андрей Павлович", *profile.FullName)
}

func TestFSPService_RegistrySearchAndVerify(t *testing.T) {
	repo := newMockRepository()
	cfg := &config.Config{
		PassThresholdRatio:    0.7,
		UpgradeThresholdRatio: 0.9,
	}
	svc := NewTestingService(repo, cfg)
	ctx := context.Background()

	// Search
	members, total, err := svc.SearchFSPRegistry(ctx, "Смирнов", "", "", 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.Equal(t, "Смирнов Александр Дмитриевич", members[0].FullName)

	// Verify
	res, err := svc.VerifyFSPMember(ctx, fsp.VerificationRequest{
		FSPID:    "FSP-RU-77-00101",
		FullName: "Смирнов Александр Дмитриевич",
	})
	require.NoError(t, err)
	assert.True(t, res.IsValid)
	assert.NotNil(t, res.Member)
}
