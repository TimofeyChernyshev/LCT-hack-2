package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/application"
	api "github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/httpx"
)

type Handlers struct {
	svc *application.Service
}

func NewHandlers(svc *application.Service) *Handlers { return &Handlers{svc: svc} }

var _ api.ServerInterface = (*Handlers)(nil)

// ---------- Healthz ----------

func (h *Handlers) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ---------- Company ----------

func (h *Handlers) GetMyCompany(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	company, err := h.svc.GetMyCompany(c.Request.Context(), u.ID)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCompanyResponse(company))
}

func (h *Handlers) UpsertMyCompany(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	var req api.CompanyInput
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	email := string(*req.ContactEmail)
	cmp, err := h.svc.UpsertMyCompany(c.Request.Context(), u.ID, application.UpsertCompanyInput{
		Name:            req.Name,
		Description:     req.Description,
		Industry:        req.Industry,
		Website:         req.Website,
		Size:            req.Size,
		ContactPerson:   req.ContactPerson,
		ContactEmail:    &email,
		ContactPhone:    req.ContactPhone,
		ContactTelegram: req.ContactTelegram,
	})
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCompanyResponse(cmp))
}

// ---------- Needs ----------

func (h *Handlers) ListMyNeeds(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	list, err := h.svc.ListMyNeeds(c.Request.Context(), u.ID)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toNeedsResponse(list))
}

func (h *Handlers) CreateNeed(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	var req api.NeedInput
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	in := application.NeedInput{
		Title:              req.Title,
		Description:        req.Description,
		CategoryID:         uuidToStringPtr(req.CategoryId),
		SpecializationID:   uuidToStringPtr(req.SpecializationId),
		GradeID:            uuidToStringPtr(req.GradeId),
		Stack:              uuidSliceFromAPIPtr(req.Stack),
		SalaryMin:          req.SalaryMin,
		SalaryMax:          req.SalaryMax,
		Location:           req.Location,
		MinExperienceYears: req.MinExperienceYears,
	}
	if req.WorkFormat != nil {
		wf := string(*req.WorkFormat)
		in.WorkFormat = &wf
	}
	if req.Status != nil {
		s := string(*req.Status)
		in.Status = &s
	}
	n, err := h.svc.CreateNeed(c.Request.Context(), u.ID, in)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toNeedResponse(n))
}

func (h *Handlers) UpdateNeed(c *gin.Context, id openapi_types.UUID) {
	u, _ := httpx.UserFromGin(c)
	var req api.NeedInput
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	in := application.NeedInput{
		Title:              req.Title,
		Description:        req.Description,
		CategoryID:         uuidToStringPtr(req.CategoryId),
		SpecializationID:   uuidToStringPtr(req.SpecializationId),
		GradeID:            uuidToStringPtr(req.GradeId),
		Stack:              uuidSliceFromAPIPtr(req.Stack),
		SalaryMin:          req.SalaryMin,
		SalaryMax:          req.SalaryMax,
		Location:           req.Location,
		MinExperienceYears: req.MinExperienceYears,
	}
	if req.WorkFormat != nil {
		wf := string(*req.WorkFormat)
		in.WorkFormat = &wf
	}
	if req.Status != nil {
		s := string(*req.Status)
		in.Status = &s
	}
	n, err := h.svc.UpdateNeed(c.Request.Context(), u.ID, id.String(), in)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toNeedResponse(n))
}

func (h *Handlers) DeleteNeed(c *gin.Context, id openapi_types.UUID) {
	u, _ := httpx.UserFromGin(c)
	if err := h.svc.DeleteNeed(c.Request.Context(), u.ID, id.String()); err != nil {
		writeDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handlers) GetNeedMatches(c *gin.Context, id openapi_types.UUID, params api.GetNeedMatchesParams) {
	u, _ := httpx.UserFromGin(c)
	limit := 20
	offset := 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}
	page, err := h.svc.GetNeedMatches(c.Request.Context(), u.ID, id.String(),
		c.GetHeader("Authorization"), application.MatchesQuery{Limit: limit, Offset: offset})
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toSearchPageResponse(page))
}

// ---------- Vacancies ----------

func (h *Handlers) ListMyVacancies(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	list, err := h.svc.ListMyVacancies(c.Request.Context(), u.ID)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVacanciesResponse(list))
}

func (h *Handlers) CreateVacancy(c *gin.Context) {
	u, _ := httpx.UserFromGin(c)
	var req api.VacancyInput
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	in := application.VacancyInput{
		Title:            req.Title,
		Description:      req.Description,
		CategoryID:       uuidToStringPtr(req.CategoryId),
		SpecializationID: uuidToStringPtr(req.SpecializationId),
		GradeID:          uuidToStringPtr(req.GradeId),
		Stack:            uuidSliceFromAPIPtr(req.Stack),
		SalaryMin:        req.SalaryMin,
		SalaryMax:        req.SalaryMax,
		Location:         req.Location,
	}
	if req.WorkFormat != nil {
		wf := string(*req.WorkFormat)
		in.WorkFormat = &wf
	}
	v, err := h.svc.CreateVacancy(c.Request.Context(), u.ID, in)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toVacancyResponse(v))
}

func (h *Handlers) UpdateVacancy(c *gin.Context, id openapi_types.UUID) {
	u, _ := httpx.UserFromGin(c)
	var req api.VacancyInput
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	in := application.VacancyInput{
		Title:            req.Title,
		Description:      req.Description,
		CategoryID:       uuidToStringPtr(req.CategoryId),
		SpecializationID: uuidToStringPtr(req.SpecializationId),
		GradeID:          uuidToStringPtr(req.GradeId),
		Stack:            uuidSliceFromAPIPtr(req.Stack),
		SalaryMin:        req.SalaryMin,
		SalaryMax:        req.SalaryMax,
		Location:         req.Location,
	}
	if req.WorkFormat != nil {
		wf := string(*req.WorkFormat)
		in.WorkFormat = &wf
	}
	v, err := h.svc.UpdateVacancy(c.Request.Context(), u.ID, id.String(), in)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVacancyResponse(v))
}

func (h *Handlers) DeleteVacancy(c *gin.Context, id openapi_types.UUID) {
	u, _ := httpx.UserFromGin(c)
	if err := h.svc.DeleteVacancy(c.Request.Context(), u.ID, id.String()); err != nil {
		writeDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handlers) PublishVacancy(c *gin.Context, id openapi_types.UUID) {
	u, _ := httpx.UserFromGin(c)
	v, err := h.svc.PublishVacancy(c.Request.Context(), u.ID, id.String())
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVacancyResponse(v))
}

func (h *Handlers) ListPublicVacancies(c *gin.Context, params api.ListPublicVacanciesParams) {
	limit := 20
	offset := 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}
	filter := application.PublicVacancyFilter{
		CategoryID: uuidToStringPtr(params.CategoryId),
		GradeID:    uuidToStringPtr(params.GradeId),
		Limit:      limit,
		Offset:     offset,
	}
	list, _, err := h.svc.ListPublicVacancies(c.Request.Context(), filter)
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVacanciesResponse(list))
}

func (h *Handlers) GetPublicVacancy(c *gin.Context, id openapi_types.UUID) {
	v, err := h.svc.GetPublicVacancy(c.Request.Context(), id.String())
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toVacancyResponse(v))
}

// ---------- Candidates (прокси) ----------

func (h *Handlers) SearchCandidates(c *gin.Context, params api.SearchCandidatesParams) {
	limit := 20
	offset := 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}
	sort := "relevance"
	if params.Sort != nil {
		sort = string(*params.Sort)
	}
	q := application.SearchQuery{
		CategoryID:         uuidToStringPtr(params.CategoryId),
		SpecializationID:   uuidToStringPtr(params.SpecializationId),
		GradeID:            uuidToStringPtr(params.GradeId),
		HasFSP:             params.HasFsp,
		MinYearsExperience: params.MinYearsExperience,
		Location:           params.Location,
		Sort:               sort,
		Limit:              limit,
		Offset:             offset,
	}
	if params.Stack != nil {
		q.Stack = uuidSliceFromAPIPtr(params.Stack)
	}
	page, err := h.svc.SearchCandidates(c.Request.Context(), c.GetHeader("Authorization"), q)
	if err != nil {
		writeError(c, http.StatusBadGateway, "search_unavailable", err.Error())
		return
	}
	c.JSON(http.StatusOK, toSearchPageResponse(page))
}

func (h *Handlers) GetCandidateCard(c *gin.Context, userId openapi_types.UUID) {
	card, err := h.svc.GetCandidateCard(c.Request.Context(), c.GetHeader("Authorization"), userId.String())
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCandidateCardResponse(card))
}

// ---------- helpers ----------

func uuidToStringPtr(u *openapi_types.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}

func uuidSliceFromAPIPtr(in *[]openapi_types.UUID) []string {
	if in == nil {
		return nil
	}
	out := make([]string, 0, len(*in))
	for _, u := range *in {
		out = append(out, u.String())
	}
	return out
}
