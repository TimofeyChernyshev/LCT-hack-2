package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/TimofeyChernyshev/LCT-hack-2/internal/interaction/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/httpx"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
)

type invitation struct {
	api.Invitation
	CompanyName     *string `json:"companyName,omitempty"`
	EmployerContact *string `json:"employerContact,omitempty"`
}

type Server struct {
	pool          *pgxpool.Pool
	candidateURL  string
	internalToken string
}

func NewServer(pool *pgxpool.Pool, candidateURL, internalToken string) *Server {
	return &Server{pool: pool, candidateURL: candidateURL, internalToken: internalToken}
}

var _ api.ServerInterface = (*Server)(nil)

func (s *Server) Healthz(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }
func (s *Server) InternalRecordReveal(c *gin.Context) { c.Status(http.StatusCreated) }

func (s *Server) CreateInvitation(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	var body struct {
		api.CreateInvitationRequest
		EmployerContact string `json:"employerContact"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Message == "" || body.SalaryMax < body.SalaryMin {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "нужны текст приглашения и вилка")
		return
	}
	employerID, _ := uuid.Parse(user.ID)
	companyID := uuid.New()
	currency := "RUB"
	if body.Currency != nil && *body.Currency != "" {
		currency = *body.Currency
	}
	item := invitation{}
	var contact *string
	if body.EmployerContact != "" {
		contact = &body.EmployerContact
	}
	err := s.pool.QueryRow(c.Request.Context(), `
		INSERT INTO invitations (employer_user_id, company_id, candidate_user_id, vacancy_id, need_id, message, salary_min, salary_max, currency, contact_channel, employer_contact)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, employer_user_id, company_id, candidate_user_id, vacancy_id, need_id, message, salary_min, salary_max, currency, contact_channel, status, created_at, employer_contact`,
		employerID, companyID, body.CandidateUserId, body.VacancyId, body.NeedId, body.Message, body.SalaryMin, body.SalaryMax, currency, body.ContactChannel, contact,
	).Scan(&item.Id, &item.EmployerUserId, &item.CompanyId, &item.CandidateUserId, &item.VacancyId, &item.NeedId, &item.Message, &item.SalaryMin, &item.SalaryMax, &item.Currency, &item.ContactChannel, &item.Status, &item.CreatedAt, &item.EmployerContact)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "приглашение не отправилось")
		return
	}
	name := "Компания"
	item.CompanyName = &name
	c.JSON(http.StatusCreated, item)
}

func (s *Server) ListMyInvitationsAsEmployer(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	id, _ := uuid.Parse(user.ID)
	items, err := s.list(c, `WHERE employer_user_id = $1`, id)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "приглашения не прочитались")
		return
	}
	c.JSON(http.StatusOK, items)
}

func (s *Server) ListIncomingInvitations(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	id, _ := uuid.Parse(user.ID)
	items, err := s.list(c, `WHERE candidate_user_id = $1`, id)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "приглашения не прочитались")
		return
	}
	c.JSON(http.StatusOK, items)
}

func (s *Server) AcceptInvitation(c *gin.Context, id openapi_types.UUID) { s.setStatus(c, id, "accepted", true) }
func (s *Server) RejectInvitation(c *gin.Context, id openapi_types.UUID) { s.setStatus(c, id, "rejected", true) }
func (s *Server) WithdrawInvitation(c *gin.Context, id openapi_types.UUID) {
	s.setStatus(c, id, "withdrawn", false)
}

func (s *Server) CreateApplication(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	var body api.CreateApplicationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "укажите вакансию")
		return
	}
	candidateID, _ := uuid.Parse(user.ID)
	var item api.Application
	status := api.ApplicationStatusSent
	now := time.Now()
	err := s.pool.QueryRow(c.Request.Context(), `
		INSERT INTO applications (candidate_user_id, vacancy_id, company_id, cover_letter)
		VALUES ($1,$2,$3,$4)
		RETURNING id, candidate_user_id, vacancy_id, company_id, cover_letter, status, created_at`,
		candidateID, body.VacancyId, uuid.New(), body.CoverLetter,
	).Scan(&item.Id, &item.CandidateUserId, &item.VacancyId, &item.CompanyId, &item.CoverLetter, &status, &item.CreatedAt)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "отклик не отправился")
		return
	}
	item.Status = &status
	item.StatusUpdatedAt = &now
	c.JSON(http.StatusCreated, item)
}

func (s *Server) ListMyApplicationsAsCandidate(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	id, _ := uuid.Parse(user.ID)
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT id, candidate_user_id, vacancy_id, company_id, cover_letter, status, created_at
		FROM applications WHERE candidate_user_id = $1 ORDER BY created_at DESC`, id)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "отклики не прочитались")
		return
	}
	defer rows.Close()
	out := []api.Application{}
	for rows.Next() {
		var item api.Application
		var status api.ApplicationStatus
		if err := rows.Scan(&item.Id, &item.CandidateUserId, &item.VacancyId, &item.CompanyId, &item.CoverLetter, &status, &item.CreatedAt); err != nil {
			httpx.GinError(c, http.StatusInternalServerError, "internal", "отклики не прочитались")
			return
		}
		item.Status = &status
		out = append(out, item)
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) ListApplicationsForVacancy(c *gin.Context, vacancyID openapi_types.UUID) {
	if _, ok := mustUser(c); !ok {
		return
	}
	c.JSON(http.StatusOK, []api.Application{})
}
func (s *Server) AcceptApplication(c *gin.Context, id openapi_types.UUID) { c.Status(http.StatusOK) }
func (s *Server) RejectApplication(c *gin.Context, id openapi_types.UUID) { c.Status(http.StatusOK) }

func (s *Server) setStatus(c *gin.Context, id openapi_types.UUID, status string, asCandidate bool) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	uid, _ := uuid.Parse(user.ID)
	column := "employer_user_id"
	if asCandidate {
		column = "candidate_user_id"
	}
	item := invitation{}
	err := s.pool.QueryRow(c.Request.Context(), `
		UPDATE invitations SET status = $3, status_updated_at = now()
		WHERE id = $1 AND `+column+` = $2
		RETURNING id, employer_user_id, company_id, candidate_user_id, vacancy_id, need_id, message, salary_min, salary_max, currency, contact_channel, status, created_at, employer_contact`,
		id, uid, status,
	).Scan(&item.Id, &item.EmployerUserId, &item.CompanyId, &item.CandidateUserId, &item.VacancyId, &item.NeedId, &item.Message, &item.SalaryMin, &item.SalaryMax, &item.Currency, &item.ContactChannel, &item.Status, &item.CreatedAt, &item.EmployerContact)
	if err != nil {
		httpx.GinError(c, http.StatusNotFound, "not_found", "приглашение не найдено")
		return
	}
	name := "Компания"
	item.CompanyName = &name
	if status == "accepted" && asCandidate {
		s.recordReveal(c.Request.Context(), item)
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) recordReveal(ctx context.Context, item invitation) {
	if s.candidateURL == "" {
		return
	}
	payload, err := json.Marshal(map[string]string{
		"candidateUserId": item.CandidateUserId.String(),
		"employerUserId":  item.EmployerUserId.String(),
		"entityId":        item.Id.String(),
		"entityType":      "invitation",
		"reason":          "invitation_accepted",
	})
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(s.candidateURL, "/")+"/internal/reveals", bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if s.internalToken != "" {
		req.Header.Set("X-Internal-Token", s.internalToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func (s *Server) list(c *gin.Context, where string, id uuid.UUID) ([]invitation, error) {
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT id, employer_user_id, company_id, candidate_user_id, vacancy_id, need_id, message, salary_min, salary_max, currency, contact_channel, status, created_at, employer_contact
		FROM invitations `+where+` ORDER BY created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []invitation{}
	for rows.Next() {
		var item invitation
		if err := rows.Scan(&item.Id, &item.EmployerUserId, &item.CompanyId, &item.CandidateUserId, &item.VacancyId, &item.NeedId, &item.Message, &item.SalaryMin, &item.SalaryMax, &item.Currency, &item.ContactChannel, &item.Status, &item.CreatedAt, &item.EmployerContact); err != nil {
			return nil, err
		}
		name := "Компания"
		item.CompanyName = &name
		out = append(out, item)
	}
	return out, rows.Err()
}

func NewRouter(si api.ServerInterface, signer jwtx.Signer, origins []string, env string) *gin.Engine {
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(cors.New(cors.Config{AllowOrigins: origins, AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Authorization", "Content-Type"}, MaxAge: 5 * time.Minute}))
	w := &api.ServerInterfaceWrapper{Handler: si, ErrorHandler: func(c *gin.Context, err error, code int) {
		c.AbortWithStatusJSON(code, gin.H{"code": "bad_request", "message": err.Error()})
	}}
	r.GET("/healthz", w.Healthz)
	r.POST("/internal/reveals", w.InternalRecordReveal)
	auth := r.Group("")
	auth.Use(httpx.GinJWTAuth(signer), httpx.GinRequireVerifiedEmail())
	auth.POST("/me/invitations", w.CreateInvitation)
	auth.GET("/me/invitations", w.ListMyInvitationsAsEmployer)
	auth.GET("/me/invitations/incoming", w.ListIncomingInvitations)
	auth.POST("/me/invitations/:id/accept", w.AcceptInvitation)
	auth.POST("/me/invitations/:id/reject", w.RejectInvitation)
	auth.POST("/me/invitations/:id/withdraw", w.WithdrawInvitation)
	auth.POST("/me/applications", w.CreateApplication)
	auth.GET("/me/applications", w.ListMyApplicationsAsCandidate)
	auth.GET("/me/vacancies/:vacancyId/applications", w.ListApplicationsForVacancy)
	auth.POST("/me/applications/:id/accept", w.AcceptApplication)
	auth.POST("/me/applications/:id/reject", w.RejectApplication)
	return r
}

func mustUser(c *gin.Context) (httpx.UserInfo, bool) {
	user, ok := httpx.UserFromGin(c)
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
	}
	return user, ok
}
