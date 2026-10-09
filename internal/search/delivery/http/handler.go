package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"TimofeyChernyshev/LCT-hack-2/internal/search/domain"
	apisearch "TimofeyChernyshev/LCT-hack-2/internal/search/infrastructure/http"
	"TimofeyChernyshev/LCT-hack-2/internal/search/repository"
	"TimofeyChernyshev/LCT-hack-2/internal/search/service"
)

type Handler struct {
	svc *service.SearchService
}

func NewHandler(svc *service.SearchService) *Handler {
	return &Handler{svc: svc}
}

// Healthz (GET /healthz)
func (h *Handler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// SearchCandidates (GET /candidates)
func (h *Handler) SearchCandidates(c *gin.Context, params apisearch.SearchCandidatesParams) {
	filter := domain.SearchFilter{}

	if params.CategoryId != nil {
		id := uuid.UUID(*params.CategoryId)
		filter.CategoryID = &id
	}
	if params.SpecializationId != nil {
		id := uuid.UUID(*params.SpecializationId)
		filter.SpecializationID = &id
	}
	if params.GradeId != nil {
		id := uuid.UUID(*params.GradeId)
		filter.GradeID = &id
	}
	if params.HasFsp != nil {
		filter.HasFSP = params.HasFsp
	}
	if params.MinYearsExperience != nil {
		exp := float64(*params.MinYearsExperience)
		filter.MinYearsExperience = &exp
	}
	if params.Location != nil {
		filter.Location = params.Location
	}
	if params.Sort != nil {
		filter.Sort = string(*params.Sort)
	}
	if params.Limit != nil {
		filter.Limit = *params.Limit
	}
	if params.Offset != nil {
		filter.Offset = *params.Offset
	}
	if params.Stack != nil {
		for _, s := range *params.Stack {
			filter.Stack = append(filter.Stack, uuid.UUID(s))
		}
	}

	page, err := h.svc.SearchCandidates(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search candidates", "details": err.Error()})
		return
	}

	respItems := make([]apisearch.SearchHit, 0, len(page.Items))
	for _, item := range page.Items {
		respItems = append(respItems, mapDomainHitToAPI(item))
	}

	c.JSON(http.StatusOK, apisearch.SearchPage{
		Items:  respItems,
		Total:  page.Total,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}

// ExplainCandidate (GET /candidates/{userId}/explain)
func (h *Handler) ExplainCandidate(c *gin.Context, userId openapi_types.UUID) {
	uid := uuid.UUID(userId)

	var targetStack []uuid.UUID
	// Optionally extract stack query params if passed
	stackParams := c.QueryArray("stack")
	for _, s := range stackParams {
		if parsed, err := uuid.Parse(s); err == nil {
			targetStack = append(targetStack, parsed)
		}
	}

	exp, err := h.svc.ExplainCandidate(c.Request.Context(), uid, targetStack)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "candidate not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to explain candidate", "details": err.Error()})
		return
	}

	factors := make([]struct {
		Contribution *float32 `json:"contribution,omitempty"`
		Detail       *string  `json:"detail,omitempty"`
		Name         *string  `json:"name,omitempty"`
		Weight       *float32 `json:"weight,omitempty"`
	}, 0, len(exp.Factors))

	for _, f := range exp.Factors {
		fContrib := f.Contribution
		fDetail := f.Detail
		fName := f.Name
		fWeight := f.Weight

		factors = append(factors, struct {
			Contribution *float32 `json:"contribution,omitempty"`
			Detail       *string  `json:"detail,omitempty"`
			Name         *string  `json:"name,omitempty"`
			Weight       *float32 `json:"weight,omitempty"`
		}{
			Contribution: &fContrib,
			Detail:       &fDetail,
			Name:         &fName,
			Weight:       &fWeight,
		})
	}

	finalScore := exp.FinalScore
	apiUID := openapi_types.UUID(exp.UserID)

	c.JSON(http.StatusOK, apisearch.ExplainResponse{
		UserId:     &apiUID,
		FinalScore: &finalScore,
		Factors:    &factors,
	})
}

// CategoryRanking (GET /categories/{categoryId}/ranking)
func (h *Handler) CategoryRanking(c *gin.Context, categoryId openapi_types.UUID, params apisearch.CategoryRankingParams) {
	catID := uuid.UUID(categoryId)
	limit := 50
	if params.Limit != nil {
		limit = *params.Limit
	}

	ranking, err := h.svc.GetCategoryRanking(c.Request.Context(), catID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get category ranking", "details": err.Error()})
		return
	}

	resp := make([]apisearch.SearchHit, 0, len(ranking))
	for _, item := range ranking {
		resp = append(resp, mapDomainHitToAPI(item))
	}

	c.JSON(http.StatusOK, resp)
}

func mapDomainHitToAPI(item domain.SearchDocHit) apisearch.SearchHit {
	hit := apisearch.SearchHit{
		UserId:               openapi_types.UUID(item.UserID),
		CategoryId:           openapi_types.UUID(item.CategoryID),
		GradeId:              openapi_types.UUID(item.GradeID),
		SpecializationId:     openapi_types.UUID(item.SpecializationID),
		Score:                item.Score,
		Reasons:              item.Reasons,
		DisplayName:          item.DisplayName,
		TestScore:            item.TestScore,
		FspAchievementsCount: item.FSPAchievementsCount,
		FspBestPlace:         item.FSPBestPlace,
		FspWeightSum:         item.FSPWeightSum,
		YearsExperience:      item.YearsExperience,
	}
	if hit.Reasons == nil {
		hit.Reasons = []string{}
	}
	return hit
}
