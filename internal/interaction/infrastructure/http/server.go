package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	employerURL   string
	internalToken string
}

func NewServer(pool *pgxpool.Pool, candidateURL, employerURL, internalToken string) *Server {
	return &Server{
		pool:          pool,
		candidateURL:  candidateURL,
		employerURL:   employerURL,
		internalToken: internalToken,
	}
}

var _ api.ServerInterface = (*Server)(nil)

func (s *Server) Healthz(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }
func (s *Server) InternalRecordReveal(c *gin.Context) { c.Status(http.StatusCreated) }

// CreateInvitation — отправка приглашения (LHH / Обратный подбор).
// Обязательная валидация: вилка ЗП в рублях (salary_min > 0, salary_max >= salary_min), описание, контакты.
func (s *Server) CreateInvitation(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	if user.Role != "employer" && user.Role != "admin" {
		httpx.GinError(c, http.StatusForbidden, "forbidden", "только работодатель может отправлять приглашения")
		return
	}

	var body struct {
		api.CreateInvitationRequest
		EmployerContact string `json:"employerContact"`
		CompanyName     string `json:"companyName"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "некорректное тело запроса")
		return
	}

	if body.CandidateUserId == uuid.Nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_candidate", "укажите кандидата")
		return
	}

	if strings.TrimSpace(body.Message) == "" {
		httpx.GinError(c, http.StatusBadRequest, "invalid_message", "описание предложения обязательно")
		return
	}

	if strings.TrimSpace(body.EmployerContact) == "" {
		httpx.GinError(c, http.StatusBadRequest, "invalid_contact", "способ связи с работодателем обязателен")
		return
	}

	// Обязательная валидация вилки зарплаты в рублях
	if body.SalaryMin <= 0 || body.SalaryMax <= 0 || body.SalaryMax < body.SalaryMin {
		httpx.GinError(c, http.StatusBadRequest, "invalid_salary", "вилка зарплаты в рублях обязательна: salary_min > 0 и salary_max >= salary_min")
		return
	}

	currency := "RUB"
	if body.Currency != nil && *body.Currency != "" {
		if *body.Currency != "RUB" {
			httpx.GinError(c, http.StatusBadRequest, "invalid_currency", "валюта вилки зарплаты должна быть в рублях (RUB)")
			return
		}
		currency = *body.Currency
	}

	employerID, err := uuid.Parse(user.ID)
	if err != nil {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "некорректный ID пользователя")
		return
	}

	companyID := uuid.New()
	companyName := "Компания"
	if strings.TrimSpace(body.CompanyName) != "" {
		companyName = strings.TrimSpace(body.CompanyName)
	}

	// Попытка получить компанию из employer service
	if s.employerURL != "" && c.GetHeader("Authorization") != "" {
		if comp, err := s.fetchEmployerCompany(c.Request.Context(), c.GetHeader("Authorization")); err == nil {
			if comp.ID != uuid.Nil {
				companyID = comp.ID
			}
			if comp.Name != "" {
				companyName = comp.Name
			}
		}
	}

	contact := strings.TrimSpace(body.EmployerContact)
	item := invitation{}
	err = s.pool.QueryRow(c.Request.Context(), `
		INSERT INTO invitations (employer_user_id, company_id, candidate_user_id, vacancy_id, need_id,
		                        message, salary_min, salary_max, currency, contact_channel,
		                        employer_contact, company_name, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,'sent')
		RETURNING id, employer_user_id, company_id, candidate_user_id, vacancy_id, need_id,
		          message, salary_min, salary_max, currency, contact_channel, status,
		          created_at, status_updated_at, employer_contact, company_name`,
		employerID, companyID, body.CandidateUserId, body.VacancyId, body.NeedId,
		body.Message, body.SalaryMin, body.SalaryMax, currency, body.ContactChannel,
		contact, companyName,
	).Scan(
		&item.Id, &item.EmployerUserId, &item.CompanyId, &item.CandidateUserId,
		&item.VacancyId, &item.NeedId, &item.Message, &item.SalaryMin, &item.SalaryMax,
		&item.Currency, &item.ContactChannel, &item.Status, &item.CreatedAt,
		&item.StatusUpdatedAt, &item.EmployerContact, &item.CompanyName,
	)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "приглашение не отправилось")
		return
	}

	if item.CompanyName == nil || *item.CompanyName == "" {
		item.CompanyName = &companyName
	}

	// Фиксация создания в истории статусов
	_, _ = s.pool.Exec(c.Request.Context(), `
		INSERT INTO interaction_status_history (entity_type, entity_id, from_status, to_status, actor_user_id)
		VALUES ('invitation', $1, NULL, 'sent', $2)`, item.Id, employerID)

	c.JSON(http.StatusCreated, item)
}

// ListMyInvitationsAsEmployer — исходящие приглашения работодателя.
func (s *Server) ListMyInvitationsAsEmployer(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	if user.Role != "employer" && user.Role != "admin" {
		httpx.GinError(c, http.StatusForbidden, "forbidden", "только работодатель")
		return
	}
	id, err := uuid.Parse(user.ID)
	if err != nil {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "некорректный ID пользователя")
		return
	}
	items, err := s.list(c, `WHERE employer_user_id = $1`, id)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "приглашения не прочитались")
		return
	}
	c.JSON(http.StatusOK, items)
}

// ListIncomingInvitations — входящие приглашения соискателя.
// Переводит статус 'sent' в 'viewed' (кандидат ознакомился с приглашением).
func (s *Server) ListIncomingInvitations(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	if user.Role != "candidate" && user.Role != "admin" {
		httpx.GinError(c, http.StatusForbidden, "forbidden", "только соискатель")
		return
	}
	id, err := uuid.Parse(user.ID)
	if err != nil {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "некорректный ID пользователя")
		return
	}

	ctx := c.Request.Context()

	// Автоматический переход sent -> viewed при просмотре списка приглашений
	rows, err := s.pool.Query(ctx, `
		UPDATE invitations
		SET status = 'viewed', status_updated_at = now()
		WHERE candidate_user_id = $1 AND status = 'sent'
		RETURNING id`, id)
	if err == nil {
		var updatedIDs []uuid.UUID
		for rows.Next() {
			var invID uuid.UUID
			if err := rows.Scan(&invID); err == nil {
				updatedIDs = append(updatedIDs, invID)
			}
		}
		rows.Close()
		for _, invID := range updatedIDs {
			_, _ = s.pool.Exec(ctx, `
				INSERT INTO interaction_status_history (entity_type, entity_id, from_status, to_status, actor_user_id)
				VALUES ('invitation', $1, 'sent', 'viewed', $2)`, invID, id)
		}
	}

	items, err := s.list(c, `WHERE candidate_user_id = $1`, id)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "приглашения не прочитались")
		return
	}
	c.JSON(http.StatusOK, items)
}

func (s *Server) AcceptInvitation(c *gin.Context, id openapi_types.UUID) {
	s.setStatus(c, id, "accepted", true)
}

func (s *Server) RejectInvitation(c *gin.Context, id openapi_types.UUID) {
	s.setStatus(c, id, "rejected", true)
}

func (s *Server) WithdrawInvitation(c *gin.Context, id openapi_types.UUID) {
	s.setStatus(c, id, "withdrawn", false)
}

func (s *Server) CreateApplication(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	if user.Role != "candidate" && user.Role != "admin" {
		httpx.GinError(c, http.StatusForbidden, "forbidden", "только соискатель")
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
	if s.pool == nil {
		c.JSON(http.StatusOK, []api.Application{})
		return
	}
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT id, candidate_user_id, vacancy_id, company_id, cover_letter, status, created_at
		FROM applications WHERE vacancy_id = $1 ORDER BY created_at DESC`, vacancyID)
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

func (s *Server) AcceptApplication(c *gin.Context, id openapi_types.UUID) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	if s.pool == nil {
		httpx.GinError(c, http.StatusNotFound, "not_found", "отклик не найден")
		return
	}
	uid, _ := uuid.Parse(user.ID)
	ctx := c.Request.Context()

	var candidateUserID uuid.UUID
	var curStatus string
	err := s.pool.QueryRow(ctx, `
		SELECT candidate_user_id, status FROM applications WHERE id = $1`, id).
		Scan(&candidateUserID, &curStatus)
	if err != nil {
		httpx.GinError(c, http.StatusNotFound, "not_found", "отклик не найден")
		return
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE applications SET status = 'accepted', status_updated_at = now() WHERE id = $1`, id)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "не удалось принять отклик")
		return
	}

	_, _ = s.pool.Exec(ctx, `
		INSERT INTO interaction_status_history (entity_type, entity_id, from_status, to_status, actor_user_id)
		VALUES ('application', $1, $2, 'accepted', $3)`, id, curStatus, uid)

	_, _ = s.pool.Exec(ctx, `
		INSERT INTO contact_reveals (candidate_user_id, employer_user_id, reason, entity_type, entity_id)
		VALUES ($1, $2, 'application_sent', 'application', $3)
		ON CONFLICT (candidate_user_id, employer_user_id, entity_type, entity_id) DO NOTHING`,
		candidateUserID, uid, id)

	c.Status(http.StatusOK)
}

func (s *Server) RejectApplication(c *gin.Context, id openapi_types.UUID) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	if s.pool == nil {
		httpx.GinError(c, http.StatusNotFound, "not_found", "отклик не найден")
		return
	}
	uid, _ := uuid.Parse(user.ID)
	ctx := c.Request.Context()

	var curStatus string
	err := s.pool.QueryRow(ctx, `
		SELECT status FROM applications WHERE id = $1`, id).
		Scan(&curStatus)
	if err != nil {
		httpx.GinError(c, http.StatusNotFound, "not_found", "отклик не найден")
		return
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE applications SET status = 'rejected', status_updated_at = now() WHERE id = $1`, id)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "не удалось отклонить отклик")
		return
	}

	_, _ = s.pool.Exec(ctx, `
		INSERT INTO interaction_status_history (entity_type, entity_id, from_status, to_status, actor_user_id)
		VALUES ('application', $1, $2, 'rejected', $3)`, id, curStatus, uid)

	c.Status(http.StatusOK)
}

// setStatus — FSM (конечный автомат) управления статусами приглашений.
// Валидирует допустимость перехода, фиксирует историю и инициирует раскрытие контактов при 'accepted'.
func (s *Server) setStatus(c *gin.Context, id openapi_types.UUID, targetStatus string, asCandidate bool) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	uid, err := uuid.Parse(user.ID)
	if err != nil {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "некорректный ID пользователя")
		return
	}

	column := "employer_user_id"
	if asCandidate {
		column = "candidate_user_id"
		if user.Role != "candidate" && user.Role != "admin" {
			httpx.GinError(c, http.StatusForbidden, "forbidden", "только соискатель")
			return
		}
	} else {
		if user.Role != "employer" && user.Role != "admin" {
			httpx.GinError(c, http.StatusForbidden, "forbidden", "только работодатель")
			return
		}
	}

	ctx := c.Request.Context()

	// 1. Получаем текущее состояние
	var cur invitation
	var curCompanyName *string
	err = s.pool.QueryRow(ctx, `
		SELECT id, employer_user_id, company_id, candidate_user_id, vacancy_id, need_id,
		       message, salary_min, salary_max, currency, contact_channel, status,
		       created_at, status_updated_at, employer_contact, company_name
		FROM invitations
		WHERE id = $1 AND `+column+` = $2`, id, uid).Scan(
		&cur.Id, &cur.EmployerUserId, &cur.CompanyId, &cur.CandidateUserId,
		&cur.VacancyId, &cur.NeedId, &cur.Message, &cur.SalaryMin, &cur.SalaryMax,
		&cur.Currency, &cur.ContactChannel, &cur.Status, &cur.CreatedAt,
		&cur.StatusUpdatedAt, &cur.EmployerContact, &curCompanyName)
	if err != nil {
		httpx.GinError(c, http.StatusNotFound, "not_found", "приглашение не найдено")
		return
	}
	if curCompanyName != nil && *curCompanyName != "" {
		cur.CompanyName = curCompanyName
	} else {
		defName := "Компания"
		cur.CompanyName = &defName
	}

	curStatus := string(cur.Status)

	// Идемпотентность
	if curStatus == targetStatus {
		c.JSON(http.StatusOK, cur)
		return
	}

	// Терминальные статусы нельзя перезаписывать
	if curStatus == "accepted" || curStatus == "rejected" || curStatus == "withdrawn" || curStatus == "expired" {
		httpx.GinError(c, http.StatusConflict, "invalid_state", "приглашение уже находится в завершённом статусе: "+curStatus)
		return
	}

	// Валидация переходов FSM
	switch targetStatus {
	case "accepted", "rejected":
		if !asCandidate {
			httpx.GinError(c, http.StatusForbidden, "forbidden", "только соискатель может принять или отклонить приглашение")
			return
		}
		if curStatus != "sent" && curStatus != "viewed" {
			httpx.GinError(c, http.StatusConflict, "invalid_transition", "невозможно перевести приглашение в статус "+targetStatus+" из статуса "+curStatus)
			return
		}
	case "withdrawn":
		if asCandidate {
			httpx.GinError(c, http.StatusForbidden, "forbidden", "только работодатель может отозвать приглашение")
			return
		}
		if curStatus != "sent" && curStatus != "viewed" {
			httpx.GinError(c, http.StatusConflict, "invalid_transition", "невозможно отозвать приглашение в статусе "+curStatus)
			return
		}
	case "viewed":
		if curStatus != "sent" {
			c.JSON(http.StatusOK, cur)
			return
		}
	default:
		httpx.GinError(c, http.StatusBadRequest, "invalid_status", "недопустимый статус "+targetStatus)
		return
	}

	item := cur
	item.Status = api.InvitationStatus(targetStatus)
	now := time.Now()
	item.StatusUpdatedAt = &now

	err = s.pool.QueryRow(ctx, `
		UPDATE invitations
		SET status = $3, status_updated_at = now()
		WHERE id = $1 AND `+column+` = $2
		RETURNING status_updated_at`,
		id, uid, targetStatus,
	).Scan(&item.StatusUpdatedAt)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "не удалось обновить статус приглашения")
		return
	}

	// Фиксация перехода в аудит-логе
	_, _ = s.pool.Exec(ctx, `
		INSERT INTO interaction_status_history (entity_type, entity_id, from_status, to_status, actor_user_id)
		VALUES ('invitation', $1, $2, $3, $4)`,
		id, curStatus, targetStatus, uid)

	// При принятии соискателем — автоматическое раскрытие контактов работодателю
	if targetStatus == "accepted" && asCandidate {
		s.recordReveal(ctx, item)
	}

	c.JSON(http.StatusOK, item)
}

func (s *Server) recordReveal(ctx context.Context, item invitation) {
	// Фиксация в contact_reveals базы данных interaction
	_, _ = s.pool.Exec(ctx, `
		INSERT INTO contact_reveals (candidate_user_id, employer_user_id, reason, entity_type, entity_id)
		VALUES ($1, $2, 'invitation_accepted', 'invitation', $3)
		ON CONFLICT (candidate_user_id, employer_user_id, entity_type, entity_id) DO NOTHING`,
		item.CandidateUserId, item.EmployerUserId, item.Id)

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
		SELECT id, employer_user_id, company_id, candidate_user_id, vacancy_id, need_id,
		       message, salary_min, salary_max, currency, contact_channel, status,
		       created_at, status_updated_at, employer_contact, company_name
		FROM invitations `+where+` ORDER BY created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []invitation{}
	for rows.Next() {
		var item invitation
		var cName *string
		if err := rows.Scan(
			&item.Id, &item.EmployerUserId, &item.CompanyId, &item.CandidateUserId,
			&item.VacancyId, &item.NeedId, &item.Message, &item.SalaryMin, &item.SalaryMax,
			&item.Currency, &item.ContactChannel, &item.Status, &item.CreatedAt,
			&item.StatusUpdatedAt, &item.EmployerContact, &cName,
		); err != nil {
			return nil, err
		}
		if cName != nil && *cName != "" {
			item.CompanyName = cName
		} else {
			defaultName := "Компания"
			item.CompanyName = &defaultName
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

type employerCompanyResp struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (s *Server) fetchEmployerCompany(ctx context.Context, authHeader string) (employerCompanyResp, error) {
	var res employerCompanyResp
	if s.employerURL == "" {
		return res, errors.New("no employer url configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(s.employerURL, "/")+"/me/company", nil)
	if err != nil {
		return res, err
	}
	req.Header.Set("Authorization", authHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return res, fmt.Errorf("employer service returned %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return res, err
	}
	return res, nil
}

func NewRouter(si api.ServerInterface, signer jwtx.Signer, origins []string, env string) *gin.Engine {
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(cors.New(cors.Config{
		AllowOrigins: origins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type", "X-Request-ID", "X-Internal-Token"},
		MaxAge:       5 * time.Minute,
	}))
	w := &api.ServerInterfaceWrapper{
		Handler: si,
		ErrorHandler: func(c *gin.Context, err error, code int) {
			c.AbortWithStatusJSON(code, gin.H{"code": "bad_request", "message": err.Error()})
		},
	}
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
	if srv, ok := si.(*Server); ok {
		auth.DELETE("/me/needs/:needId/responses", srv.DeleteNeedResponses)
	}
	return r
}

func (s *Server) DeleteNeedResponses(c *gin.Context) {
	user, ok := mustUser(c)
	if !ok {
		return
	}
	if user.Role != "employer" && user.Role != "admin" {
		httpx.GinError(c, http.StatusForbidden, "forbidden", "только работодатель")
		return
	}
	needID, err := uuid.Parse(c.Param("needId"))
	if err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_id", "некорректная потребность")
		return
	}
	uid, _ := uuid.Parse(user.ID)
	ctx := c.Request.Context()
	if _, err := s.pool.Exec(ctx, `
		DELETE FROM contact_reveals
		WHERE entity_type = 'invitation'
		  AND entity_id IN (
		    SELECT id FROM invitations WHERE need_id = $1 AND employer_user_id = $2
		  )`, needID, uid); err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "не удалось удалить реакции")
		return
	}
	if _, err := s.pool.Exec(ctx, `
		DELETE FROM interaction_status_history
		WHERE entity_type = 'invitation'
		  AND entity_id IN (
		    SELECT id FROM invitations WHERE need_id = $1 AND employer_user_id = $2
		  )`, needID, uid); err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "не удалось удалить реакции")
		return
	}
	if _, err := s.pool.Exec(ctx, `
		DELETE FROM invitations WHERE need_id = $1 AND employer_user_id = $2`, needID, uid); err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "не удалось удалить реакции")
		return
	}
	c.Status(http.StatusNoContent)
}

func mustUser(c *gin.Context) (httpx.UserInfo, bool) {
	user, ok := httpx.UserFromGin(c)
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
	}
	return user, ok
}
