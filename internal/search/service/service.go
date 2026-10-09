package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"TimofeyChernyshev/LCT-hack-2/internal/search/domain"
	"TimofeyChernyshev/LCT-hack-2/internal/search/infrastructure/config"
	"TimofeyChernyshev/LCT-hack-2/internal/search/repository"
	"TimofeyChernyshev/LCT-hack-2/pkg/ranking"
)

type SearchService struct {
	repo repository.Repository
	cfg  *config.Config
}

func NewSearchService(repo repository.Repository, cfg *config.Config) *SearchService {
	return &SearchService{
		repo: repo,
		cfg:  cfg,
	}
}

// SearchCandidates searches the candidate bank with multi-factor ranking and explanation
func (s *SearchService) SearchCandidates(ctx context.Context, filter domain.SearchFilter) (*domain.SearchPage, error) {
	docs, total, err := s.repo.Search(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("search repo: %w", err)
	}

	hits := make([]domain.SearchDocHit, 0, len(docs))
	hasTargetStack := len(filter.Stack) > 0

	for _, doc := range docs {
		hit := s.buildSearchHit(doc, filter.Stack)
		hits = append(hits, hit)
	}

	// If employer specified a specific stack, re-sort hits by dynamic stack match relevance
	if hasTargetStack && (filter.Sort == "" || filter.Sort == "relevance") {
		sort.SliceStable(hits, func(i, j int) bool {
			return hits[i].Score > hits[j].Score
		})
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = s.cfg.DefaultLimit
	}

	return &domain.SearchPage{
		Items:  hits,
		Total:  total,
		Limit:  limit,
		Offset: filter.Offset,
	}, nil
}

// ExplainCandidate returns full Explainable AI factor breakdown and justification for a candidate
func (s *SearchService) ExplainCandidate(ctx context.Context, userID uuid.UUID, targetStack []uuid.UUID) (*domain.ExplainResponse, error) {
	doc, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get candidate doc: %w", err)
	}

	features := s.mapDocToFeatures(*doc)
	percentile, _ := s.repo.CalculateTestPercentile(ctx, doc.CategoryID, features.TestScore)
	features.TestPercentile = percentile

	rankingParams := ranking.RankingParams{
		TargetStack:     targetStack,
		AdaptiveWeights: true,
	}

	ranked := ranking.Calculate(features, rankingParams)

	return &domain.ExplainResponse{
		UserID:     userID,
		FinalScore: ranked.FinalScore,
		Factors:    ranked.Factors,
		Reasons:    ranked.Reasons,
		Summary:    ranked.Explanation,
	}, nil
}

// GetCategoryRanking returns candidates in category strictly ranked by profile strength
func (s *SearchService) GetCategoryRanking(ctx context.Context, categoryID uuid.UUID, limit int) ([]domain.SearchDocHit, error) {
	docs, err := s.repo.GetCategoryRanking(ctx, categoryID, limit)
	if err != nil {
		return nil, fmt.Errorf("repo category ranking: %w", err)
	}

	hits := make([]domain.SearchDocHit, 0, len(docs))
	for _, doc := range docs {
		hit := s.buildSearchHit(doc, nil)
		hits = append(hits, hit)
	}

	return hits, nil
}

// IndexCandidate calculates scores and indexes or updates a candidate document in search read-model
func (s *SearchService) IndexCandidate(ctx context.Context, doc *domain.CandidateSearchDoc) error {
	features := s.mapDocToFeatures(*doc)
	percentile, _ := s.repo.CalculateTestPercentile(ctx, doc.CategoryID, features.TestScore)
	features.TestPercentile = percentile

	ranked := ranking.Calculate(features, ranking.RankingParams{})

	doc.FSPScore = float64(ranked.FSPScore)
	doc.ActivityScore = float64(ranked.ActivityScore)
	doc.CalculatedScore = float64(ranked.FinalScore)
	doc.Explanation = ranked.Explanation
	doc.Reasons = ranked.Reasons
	if doc.UpdatedAt.IsZero() {
		doc.UpdatedAt = time.Now()
	}

	return s.repo.Upsert(ctx, doc)
}

func (s *SearchService) buildSearchHit(doc domain.CandidateSearchDoc, targetStack []uuid.UUID) domain.SearchDocHit {
	var score float32
	var reasons []string
	var explanation string
	var factors []ranking.ScoreFactor

	if len(targetStack) > 0 {
		// Re-calculate dynamically based on employer's requested stack
		features := s.mapDocToFeatures(doc)
		ranked := ranking.Calculate(features, ranking.RankingParams{
			TargetStack:     targetStack,
			AdaptiveWeights: true,
		})
		score = ranked.FinalScore
		reasons = ranked.Reasons
		explanation = ranked.Explanation
		factors = ranked.Factors
	} else {
		// Use stored pre-calculated values
		score = float32(doc.CalculatedScore)
		reasons = doc.Reasons
		explanation = doc.Explanation
		if score <= 0 {
			features := s.mapDocToFeatures(doc)
			ranked := ranking.Calculate(features, ranking.RankingParams{})
			score = ranked.FinalScore
			reasons = ranked.Reasons
			explanation = ranked.Explanation
			factors = ranked.Factors
		}
	}

	var displayName *string
	if doc.DisplayName != "" {
		dn := doc.DisplayName
		displayName = &dn
	}

	var testScore *float32
	if doc.TestScore != nil {
		ts := float32(*doc.TestScore)
		testScore = &ts
	}

	var yearsExp *float32
	if doc.YearsExperience != nil {
		ye := float32(*doc.YearsExperience)
		yearsExp = &ye
	}

	fspCount := doc.FSPAchievementsCount
	fspWeight := doc.FSPWeightSum

	return domain.SearchDocHit{
		UserID:               doc.UserID,
		DisplayName:          displayName,
		CategoryID:           doc.CategoryID,
		GradeID:              doc.GradeID,
		SpecializationID:     doc.SpecializationID,
		TestScore:            testScore,
		FSPAchievementsCount: &fspCount,
		FSPBestPlace:         doc.FSPBestPlace,
		FSPWeightSum:         &fspWeight,
		YearsExperience:      yearsExp,
		Score:                score,
		Reasons:              reasons,
		Explanation:          explanation,
		Factors:              factors,
	}
}

func (s *SearchService) mapDocToFeatures(doc domain.CandidateSearchDoc) ranking.CandidateFeatures {
	testScore := 0.0
	if doc.TestScore != nil {
		testScore = *doc.TestScore
	}

	return ranking.CandidateFeatures{
		UserID:               doc.UserID,
		DisplayName:          doc.DisplayName,
		CategoryID:           doc.CategoryID,
		SpecializationID:     doc.SpecializationID,
		SpecializationName:   doc.SpecializationName,
		GradeID:              doc.GradeID,
		GradeName:            doc.GradeName,
		GradeRank:            doc.GradeRank,
		TestScore:            testScore,
		HasFSP:               doc.HasFSP,
		FSPScore:             doc.FSPScore,
		FSPRating:            doc.FSPRating,
		SportsRank:           doc.SportsRank,
		FSPAchievementsCount: doc.FSPAchievementsCount,
		FSPBestPlace:         doc.FSPBestPlace,
		FSPWeightSum:         doc.FSPWeightSum,
		FSPHighlights:        doc.FSPHighlights,
		CandidateStack:       doc.Stack,
		YearsExperience:      doc.YearsExperience,
		PeriodicTasksSolved:  doc.PeriodicTasksSolved,
		LastActiveAt:         &doc.LastActiveAt,
	}
}
