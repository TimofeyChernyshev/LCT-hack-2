package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/fsp"
)

// LinkFSP links candidate's account to an FSP registry member ID, enriching profile with achievements
func (s *TestingService) LinkFSP(ctx context.Context, userID uuid.UUID, fspMemberID string, source string) (*domain.CandidateFSPProfile, error) {
	normID := strings.TrimSpace(fspMemberID)
	if normID == "" {
		return nil, errors.New("fspMemberId is required")
	}

	if source == "" {
		source = "manual"
	}

	// 1. Fetch member details and achievements from FSP registry
	member, err := s.fspClient.GetMember(ctx, normID)
	if err != nil {
		if errors.Is(err, fsp.ErrMemberNotFound) {
			return nil, ErrFSPMemberNotFound
		}
		return nil, fmt.Errorf("lookup fsp member: %w", err)
	}

	// 2. Calculate score and explanation
	scoreRes := fsp.CalculateScore(member)

	// 3. Construct enriched profile
	now := time.Now()
	rankStr := string(member.SportsRank)
	disciplineStr := string(member.Discipline)

	profile := domain.CandidateFSPProfile{
		UserID:             userID,
		FSPMemberID:        &member.FSPID,
		FullName:           &member.FullName,
		SportsRank:         &rankStr,
		FSPRating:          member.Rating,
		Region:             &member.Region,
		Discipline:         &disciplineStr,
		HasFSP:             true,
		FSPScore:           scoreRes.Score,
		FSPWeightSum:       scoreRes.WeightSum,
		AchievementsCount:  scoreRes.AchievementsCount,
		BestPlace:          scoreRes.BestPlace,
		VerificationSource: source,
		LinkedAt:           &now,
		Explanation:        scoreRes.Explanation,
		CreatedAt:          now,
		UpdatedAt:          now,
		Achievements:       member.Achievements,
	}

	// 4. Persist to database
	if err := s.repo.SaveCandidateFSP(ctx, &profile); err != nil {
		return nil, fmt.Errorf("save candidate fsp profile: %w", err)
	}

	// 5. Notify candidate microservice asynchronously if configured
	go s.notifyCandidateFSP(userID, &profile)

	return &profile, nil
}

// GetCandidateFSP returns the candidate's FSP profile.
// Critical Edge Case (No-FSP): Candidates without linked FSP receive a neutral profile with score = 0,
// empty achievements, and a transparent explanation string. No errors are returned.
func (s *TestingService) GetCandidateFSP(ctx context.Context, userID uuid.UUID) (*domain.CandidateFSPProfile, error) {
	p, err := s.repo.GetCandidateFSP(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get candidate fsp: %w", err)
	}

	if p != nil && p.HasFSP {
		return p, nil
	}

	// Return neutral No-FSP state
	neutral := domain.CandidateFSPProfile{
		UserID:             userID,
		HasFSP:             false,
		FSPMemberID:        nil,
		FullName:           nil,
		SportsRank:         nil,
		FSPRating:          0,
		Region:             nil,
		Discipline:         nil,
		FSPScore:           0.0,
		FSPWeightSum:       0,
		AchievementsCount:  0,
		BestPlace:          nil,
		VerificationSource: "none",
		LinkedAt:           nil,
		Explanation:        fsp.NoFSPExplanation,
		Achievements:       []fsp.Achievement{},
	}
	return &neutral, nil
}

// UnlinkFSP detaches FSP registry connection and returns neutral state
func (s *TestingService) UnlinkFSP(ctx context.Context, userID uuid.UUID) (*domain.CandidateFSPProfile, error) {
	if err := s.repo.UnlinkCandidateFSP(ctx, userID); err != nil {
		return nil, fmt.Errorf("unlink candidate fsp: %w", err)
	}

	neutral := domain.CandidateFSPProfile{
		UserID:             userID,
		HasFSP:             false,
		FSPMemberID:        nil,
		FullName:           nil,
		SportsRank:         nil,
		FSPRating:          0,
		Region:             nil,
		Discipline:         nil,
		FSPScore:           0.0,
		FSPWeightSum:       0,
		AchievementsCount:  0,
		BestPlace:          nil,
		VerificationSource: "none",
		LinkedAt:           nil,
		Explanation:        fsp.NoFSPExplanation,
		Achievements:       []fsp.Achievement{},
	}

	// Notify candidate microservice of unlinking
	go s.notifyCandidateFSP(userID, &neutral)

	return &neutral, nil
}

// SyncKeycloakFSP extracts FSP claims from Keycloak token and automatically links if present
func (s *TestingService) SyncKeycloakFSP(ctx context.Context, userID uuid.UUID, claims map[string]any) (*domain.CandidateFSPProfile, error) {
	fspID, found := s.keycloakHelper.ExtractFSPID(claims)
	if !found || strings.TrimSpace(fspID) == "" {
		// No Keycloak FSP claim found, return current state
		return s.GetCandidateFSP(ctx, userID)
	}

	return s.LinkFSP(ctx, userID, fspID, "keycloak")
}

// SearchFSPRegistry searches members in FSP registry
func (s *TestingService) SearchFSPRegistry(ctx context.Context, query string, rank fsp.SportsRank, region string, limit, offset int) ([]fsp.Member, int, error) {
	return s.fspClient.SearchMembers(ctx, query, rank, region, limit, offset)
}

// GetFSPRegistryMember retrieves raw registry record for member
func (s *TestingService) GetFSPRegistryMember(ctx context.Context, fspID string) (*fsp.Member, error) {
	return s.fspClient.GetMember(ctx, fspID)
}

// VerifyFSPMember verifies member credentials against registry
func (s *TestingService) VerifyFSPMember(ctx context.Context, req fsp.VerificationRequest) (*fsp.VerificationResult, error) {
	return s.fspClient.VerifyMember(ctx, req)
}

// notifyCandidateFSP pushes updated FSP status to candidate microservice if reachable
func (s *TestingService) notifyCandidateFSP(userID uuid.UUID, profile *domain.CandidateFSPProfile) {
	if s.cfg.CandidateURL == "" {
		return
	}

	url := fmt.Sprintf("%s/internal/candidates/%s/fsp", s.cfg.CandidateURL, userID.String())
	body, err := json.Marshal(profile)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}
