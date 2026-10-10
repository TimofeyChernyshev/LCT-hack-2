package http

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/search/domain"
	apisearch "github.com/TimofeyChernyshev/LCT-hack-2/internal/search/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/search/repository"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/search/service"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/resume"
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

// ExportCandidatePDF (GET /candidates/:userId/export-pdf or POST /candidates/export-pdf)
func (h *Handler) ExportCandidatePDF(c *gin.Context) {
	var profile resume.CandidateExportProfile

	// Check if userId is passed in URL path
	uidStr := c.Param("userId")
	if uidStr != "" {
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}

		doc, err := h.svc.GetCandidateProfile(c.Request.Context(), uid)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "candidate not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load candidate profile", "details": err.Error()})
			return
		}

		testScore := 80.0
		if doc.TestScore != nil {
			testScore = *doc.TestScore
		}
		yearsExp := 3.0
		if doc.YearsExperience != nil {
			yearsExp = *doc.YearsExperience
		}
		sportsRank := doc.SportsRank
		location := "Москва"
		if doc.Location != nil && *doc.Location != "" {
			location = *doc.Location
		}

		unlocked := c.Query("unlocked") == "true"

		profile = resume.CandidateExportProfile{
			UserID:             doc.UserID,
			FullName:           doc.DisplayName,
			Headline:           fmt.Sprintf("%s %s Developer", doc.GradeName, doc.SpecializationName),
			SpecializationName: doc.SpecializationName,
			GradeName:          doc.GradeName,
			GradeRank:          doc.GradeRank,
			Location:           location,
			TestScore:          testScore,
			TestPercentile:     95.0,
			HasFSP:             doc.HasFSP,
			SportsRank:         sportsRank,
			FSPRating:          doc.FSPRating,
			FSPAchievements:    doc.FSPHighlights,
			YearsExperience:    yearsExp,
			MaskContacts:       !unlocked,
			Email:              fmt.Sprintf("candidate-%s@fsp-platform.ru", doc.UserID.String()[:8]),
			Phone:              "+7 (999) 000-00-00",
			Telegram:           fmt.Sprintf("@candidate_%s", doc.UserID.String()[:8]),
			GeneratedAt:        time.Now(),
		}
	} else if c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&profile)
	}

	if profile.FullName == "" {
		profile.FullName = "Александр Дмитриевич Смирнов"
		profile.Headline = "Senior Backend Developer"
		profile.SpecializationName = "Backend"
		profile.GradeName = "Senior"
		profile.Location = "Москва"
		profile.TestScore = 95.0
		profile.TestPercentile = 98.0
		profile.HasFSP = true
		profile.SportsRank = "Мастер спорта"
		profile.Skills = []string{"Go", "PostgreSQL", "Redis", "Docker", "Kubernetes"}
		profile.YearsExperience = 6.0
		profile.MaskContacts = true
	}
	if profile.UserID == uuid.Nil {
		profile.UserID = uuid.New()
	}
	if profile.GeneratedAt.IsZero() {
		profile.GeneratedAt = time.Now()
	}

	pdfBytes, err := resume.GenerateCandidatePDF(profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate candidate PDF: " + err.Error()})
		return
	}

	filename := fmt.Sprintf("fsp_profile_%s.pdf", profile.UserID.String()[:8])
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// ParseResumePDF (POST /resumes/parse-pdf or POST /me/resumes/upload)
func (h *Handler) ParseResumePDF(c *gin.Context) {
	var fileBytes []byte

	// 1. Multipart file
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		file, _, err = c.Request.FormFile("resume")
	}
	if err == nil {
		defer file.Close()
		data, readErr := io.ReadAll(file)
		if readErr == nil && len(data) > 0 {
			fileBytes = data
		}
	}

	// 2. JSON payload
	if len(fileBytes) == 0 {
		var jsonReq struct {
			PDFBase64 string `json:"pdf_base64"`
			Text      string `json:"text"`
			Content   string `json:"content"`
		}
		if err := c.ShouldBindJSON(&jsonReq); err == nil {
			if jsonReq.PDFBase64 != "" {
				if dec, decErr := base64.StdEncoding.DecodeString(jsonReq.PDFBase64); decErr == nil {
					fileBytes = dec
				}
			} else if jsonReq.Text != "" {
				fileBytes = []byte(jsonReq.Text)
			} else if jsonReq.Content != "" {
				fileBytes = []byte(jsonReq.Content)
			}
		}
	}

	// 3. Fallback raw body
	if len(fileBytes) == 0 && c.Request.Body != nil {
		raw, _ := io.ReadAll(c.Request.Body)
		if len(raw) > 0 {
			fileBytes = raw
		}
	}

	if len(fileBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "empty resume payload: provide 'file' via multipart form, raw PDF body, or JSON with 'text'/'pdf_base64'",
		})
		return
	}

	parsed, err := resume.ExtractAndParse(fileBytes)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "failed to extract/parse resume PDF: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, parsed)
}
