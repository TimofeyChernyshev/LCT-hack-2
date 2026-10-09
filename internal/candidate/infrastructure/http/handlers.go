package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/application"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	api "github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/httpx"
)

type Handlers struct {
	svc           *application.Service
	internalToken string
}

func NewHandlers(svc *application.Service, internalToken string) *Handlers {
	return &Handlers{svc: svc, internalToken: internalToken}
}

var _ api.ServerInterface = (*Handlers)(nil)

// ---------- Healthz ----------

func (h *Handlers) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ---------- Profile ----------

func (h *Handlers) GetMyProfile(c *gin.Context) {
	u, ok := httpx.UserFromGin(c)
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
		return
	}
	p, err := h.svc.GetMyProfile(c.Request.Context(), u.ID)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProfileResponse(p))
}

func (h *Handlers) UpdateMyProfile(c *gin.Context) {
	u, ok := httpx.UserFromGin(c)
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
		return
	}
	var req api.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	in := application.UpdateProfileInput{
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		MiddleName:      req.MiddleName,
		Headline:        req.Headline,
		About:           req.About,
		Location:        req.Location,
		YearsExperience: req.YearsExperience,
		SalaryMin:       req.SalaryMin,
		SalaryMax:       req.SalaryMax,
		SalaryCurrency:  req.SalaryCurrency,
	}
	if req.SoftSkills != nil {
		in.SoftSkills = *req.SoftSkills
	}
	p, err := h.svc.UpdateMyProfile(c.Request.Context(), u.ID, in)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProfileResponse(p))
}

// ---------- Contacts ----------

func (h *Handlers) GetMyContacts(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	cnt, err := h.svc.GetMyContacts(c.Request.Context(), u.ID)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toContactsResponse(cnt))
}

func (h *Handlers) ReplaceMyContacts(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	var req api.CandidateContacts
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	cnt := &domain.Contacts{
		Phone:    req.Phone,
		Telegram: req.Telegram,
		GitHub:   req.Github,
		LinkedIn: req.Linkedin,
		Website:  req.Website,
	}
	if req.Email != nil {
		cnt.Email = string(*req.Email)
	}
	if err := h.svc.ReplaceMyContacts(c.Request.Context(), u.ID, cnt); err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toContactsResponse(cnt))
}

// ---------- Visibility ----------

func (h *Handlers) GetMyVisibility(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	v, err := h.svc.GetMyVisibility(c.Request.Context(), u.ID)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVisibilityResponse(v))
}

func (h *Handlers) UpdateMyVisibility(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	var req api.Visibility
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	v := domain.DefaultVisibility()
	if req.Contacts != nil {
		v.Contacts = *req.Contacts
	}
	if req.Links != nil {
		v.Links = *req.Links
	}
	if req.Fsp != nil {
		v.FSP = *req.Fsp
	}
	if req.Experience != nil {
		v.Experience = *req.Experience
	}
	if req.Resume != nil {
		v.Resume = *req.Resume
	}
	if req.Salary != nil {
		v.Salary = *req.Salary
	}
	if req.SoftSkills != nil {
		v.SoftSkills = *req.SoftSkills
	}
	if err := h.svc.UpdateMyVisibility(c.Request.Context(), u.ID, &v); err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ---------- Employer view ----------

func (h *Handlers) GetCandidateForEmployer(c *gin.Context, userId openapi_types.UUID) {
	ctx := c.Request.Context()

	actor, ok := httpx.UserFromGin(c)
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
		return
	}
	if actor.Role != "employer" && actor.Role != "admin" {
		httpx.GinError(c, http.StatusForbidden, "forbidden", "employers only")
		return
	}

	view, err := h.svc.BuildCardView(ctx, userId.String(), actor.ID)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			httpx.GinError(c, http.StatusNotFound, "not_found", "profile not found")
			return
		}
		httpx.GinError(c, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}
	c.JSON(http.StatusOK, toCardResponse(view))
}

// ---------- Internal ----------

func (h *Handlers) InternalRecordReveal(c *gin.Context, _ api.InternalRecordRevealParams) {
	if h.internalToken != "" && c.GetHeader("X-Internal-Token") != h.internalToken {
		httpx.GinError(c, http.StatusForbidden, "forbidden", "invalid internal token")
		return
	}
	var req api.InternalRecordRevealJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	candID := req.CandidateUserId.String()
	empID := req.EmployerUserId.String()
	entityID := req.EntityId.String()

	err := h.svc.RecordReveal(c.Request.Context(), application.RecordRevealInput{
		CandidateUserID: candID,
		EmployerUserID:  empID,
		EntityType:      domain.RevealEntityType(req.EntityType),
		EntityID:        entityID,
		Reason:          domain.RevealReason(req.Reason),
	})
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------- Resumes ----------

func (h *Handlers) ListResumes(c *gin.Context) {
	u, ok := httpx.UserFromGin(c)
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
		return
	}
	list, err := h.svc.ListResumes(c.Request.Context(), u.ID)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResumesResponse(list))
}

func (h *Handlers) CreateResume(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	var req api.ResumeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	isPrimary := false
	if req.IsPrimary != nil {
		isPrimary = *req.IsPrimary
	}
	r, err := h.svc.CreateResume(c.Request.Context(), u.ID, application.CreateResumeInput{
		Title:     req.Title,
		Content:   req.Content,
		IsPrimary: isPrimary,
	})
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toResumeResponse(r))
}

func (h *Handlers) UploadResumePDF(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	file, err := c.FormFile("file")
	if err != nil {
		httpx.GinError(c, http.StatusBadRequest, "no_file", "file is required")
		return
	}
	r, err := h.svc.UploadResumePDF(c.Request.Context(), u.ID, file)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toResumeResponse(r))
}

// ---------- Experiences ----------

func (h *Handlers) ListExperiences(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	list, err := h.svc.ListExperiences(c.Request.Context(), u.ID)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toExperiencesResponse(list))
}

func (h *Handlers) AddExperience(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	var req api.ExperienceInput
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}

	e, err := h.svc.AddExperience(c.Request.Context(), u.ID, application.AddExperienceInput{
		Company:     req.Company,
		Position:    req.Position,
		StartedAt:   req.StartedAt.Time,
		EndedAt:     req.EndedAt.Time,
		Description: req.Description,
	})
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toExperienceResponse(e))
}

// ---------- Technologies ----------

func (h *Handlers) ReplaceTechnologies(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	var req []api.CandidateTechnology
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	in := make([]application.TechnologyInput, 0, len(req))
	for _, t := range req {
		in = append(in, application.TechnologyInput{
			TechnologyID: t.TechnologyId.String(),
			Level:        t.Level,
		})
	}
	if err := h.svc.ReplaceTechnologies(c.Request.Context(), u.ID, in); err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ---------- Category ----------

func (h *Handlers) GetMyCategory(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	st, err := h.svc.GetMyCategory(c.Request.Context(), u.ID)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCategoryStateResponse(st))
}

// ---------- FSP ----------

func (h *Handlers) GetMyFSP(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	st, err := h.svc.GetMyFSP(c.Request.Context(), u.ID)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toFSPStateResponse(st))
}

func (h *Handlers) LinkFSP(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	var req struct {
		FspMemberId string `json:"fspMemberId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	if err := h.svc.LinkFSP(c.Request.Context(), u.ID, req.FspMemberId); err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
