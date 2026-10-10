package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/httpx"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
)

type Server struct{ pool *pgxpool.Pool }

func NewServer(pool *pgxpool.Pool) *Server { return &Server{pool: pool} }

var _ api.ServerInterface = (*Server)(nil)

func (s *Server) Healthz(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }

func (s *Server) GetMyCompany(c *gin.Context) {
	owner, ok := owner(c)
	if !ok {
		return
	}
	item, err := s.company(c, owner)
	if err != nil {
		httpx.GinError(c, http.StatusNotFound, "not_found", "компания ещё не создана")
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) UpsertMyCompany(c *gin.Context) {
	owner, ok := owner(c)
	if !ok {
		return
	}
	var body api.CompanyInput
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "укажите название компании")
		return
	}
	var item api.Company
	err := s.pool.QueryRow(c.Request.Context(), `
		INSERT INTO companies (owner_user_id, name, description, industry, website, size)
		VALUES ($1,$2,NULLIF($3,''),NULLIF($4,''),NULLIF($5,''),NULLIF($6,''))
		ON CONFLICT (owner_user_id) DO UPDATE SET
		  name = EXCLUDED.name,
		  description = EXCLUDED.description,
		  industry = EXCLUDED.industry,
		  website = EXCLUDED.website,
		  size = EXCLUDED.size,
		  updated_at = now()
		RETURNING id, name, description, industry, website, size`,
		owner, body.Name, str(body.Description), str(body.Industry), str(body.Website), str(body.Size),
	).Scan(&item.Id, &item.Name, &item.Description, &item.Industry, &item.Website, &item.Size)
	if err != nil {
		slog.Error("upsert company", "err", err)
		httpx.GinError(c, http.StatusInternalServerError, "internal", "не удалось сохранить компанию")
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) ListMyNeeds(c *gin.Context) {
	owner, ok := owner(c)
	if !ok {
		return
	}
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT n.id, n.title, n.description, n.category_id, n.specialization_id, n.grade_id,
		       n.stack, n.salary_min, n.salary_max, n.currency, n.work_format, n.status
		FROM employer_needs n JOIN companies c ON c.id = n.company_id
		WHERE c.owner_user_id = $1 ORDER BY n.created_at DESC`, owner)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "потребности не прочитались")
		return
	}
	defer rows.Close()
	out := []api.Need{}
	for rows.Next() {
		item, err := readNeed(rows)
		if err != nil {
			httpx.GinError(c, http.StatusInternalServerError, "internal", "потребности не прочитались")
			return
		}
		out = append(out, item)
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) CreateNeed(c *gin.Context) {
	owner, ok := owner(c)
	if !ok {
		return
	}
	body, ok := bindNeed(c)
	if !ok {
		return
	}
	company, err := s.company(c, owner)
	if err != nil {
		httpx.GinError(c, http.StatusConflict, "company_required", "сначала сохраните компанию")
		return
	}
	var item api.Need
	var stack []uuid.UUID
	var format *string
	var status string
	err = s.pool.QueryRow(c.Request.Context(), `
		INSERT INTO employer_needs (company_id, title, description, category_id, specialization_id, grade_id, stack, salary_min, salary_max, work_format, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, title, description, category_id, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status`,
		company.Id, body.Title, body.Description, body.CategoryId, body.SpecializationId, body.GradeId, uuids(body.Stack), body.SalaryMin, body.SalaryMax, formatOfNeed(body.WorkFormat), statusOf(body.Status),
	).Scan(&item.Id, &item.Title, &item.Description, &item.CategoryId, &item.SpecializationId, &item.GradeId, &stack, &item.SalaryMin, &item.SalaryMax, &item.Currency, &format, &status)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "не удалось сохранить потребность")
		return
	}
	item.Stack = uuidOut(stack)
	item.Status = api.NeedStatus(status)
	if format != nil {
		value := api.NeedWorkFormat(*format)
		item.WorkFormat = &value
	}
	c.JSON(http.StatusCreated, item)
}

func (s *Server) UpdateNeed(c *gin.Context, id openapi_types.UUID) { s.writeMissing(c, id) }
func (s *Server) DeleteNeed(c *gin.Context, id openapi_types.UUID) {
	owner, ok := owner(c)
	if !ok {
		return
	}
	tag, err := s.pool.Exec(c.Request.Context(), `
		DELETE FROM employer_needs n USING companies c
		WHERE n.id = $1 AND n.company_id = c.id AND c.owner_user_id = $2`, id, owner)
	if err != nil || tag.RowsAffected() == 0 {
		httpx.GinError(c, http.StatusNotFound, "not_found", "потребность не найдена")
		return
	}
	c.Status(http.StatusNoContent)
}
func (s *Server) GetNeedMatches(c *gin.Context, id openapi_types.UUID, _ api.GetNeedMatchesParams) {
	if _, ok := owner(c); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": []api.MatchCard{}, "total": 0})
}
func (s *Server) ListMyVacancies(c *gin.Context) {
	owner, ok := owner(c)
	if !ok {
		return
	}
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT v.id, v.company_id, v.title, v.description, v.category_id, v.stack,
		       v.salary_min, v.salary_max, v.work_format, v.status, v.published_at, v.created_at
		FROM vacancies v JOIN companies c ON c.id = v.company_id
		WHERE c.owner_user_id = $1 ORDER BY v.created_at DESC`, owner)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "вакансии не прочитались")
		return
	}
	defer rows.Close()
	out := []api.Vacancy{}
	for rows.Next() {
		item, err := readVacancy(rows)
		if err != nil {
			httpx.GinError(c, http.StatusInternalServerError, "internal", "ошибка чтения вакансии")
			return
		}
		out = append(out, item)
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) CreateVacancy(c *gin.Context) {
	owner, ok := owner(c)
	if !ok {
		return
	}
	var body api.VacancyInput
	if err := c.ShouldBindJSON(&body); err != nil || body.Title == "" || body.Description == "" || body.SalaryMax < body.SalaryMin {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "проверьте название, описание и вилку зарплаты")
		return
	}
	company, err := s.company(c, owner)
	if err != nil {
		httpx.GinError(c, http.StatusConflict, "company_required", "сначала сохраните компанию")
		return
	}
	var id uuid.UUID
	var pubAt *time.Time
	var createdAt time.Time
	var stack []uuid.UUID
	var format *string
	var status string
	err = s.pool.QueryRow(c.Request.Context(), `
		INSERT INTO vacancies (company_id, title, description, category_id, stack, salary_min, salary_max, work_format, status, published_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'published',now(),now(),now())
		RETURNING id, title, description, category_id, stack, salary_min, salary_max, work_format, status, published_at, created_at`,
		company.Id, body.Title, body.Description, body.CategoryId, uuids(body.Stack), body.SalaryMin, body.SalaryMax, formatOfVacancy(body.WorkFormat),
	).Scan(&id, &body.Title, &body.Description, &body.CategoryId, &stack, &body.SalaryMin, &body.SalaryMax, &format, &status, &pubAt, &createdAt)
	if err != nil {
		slog.Error("create vacancy", "err", err)
		httpx.GinError(c, http.StatusInternalServerError, "internal", "не удалось сохранить вакансию")
		return
	}
	st := api.VacancyStatus(status)
	item := api.Vacancy{
		Id:          &id,
		CompanyId:   &company.Id,
		Title:       body.Title,
		Description: body.Description,
		CategoryId:  body.CategoryId,
		Stack:       uuidOut(stack),
		SalaryMin:   body.SalaryMin,
		SalaryMax:   body.SalaryMax,
		Status:      &st,
		PublishedAt: pubAt,
		CreatedAt:   &createdAt,
	}
	if format != nil {
		f := api.VacancyWorkFormat(*format)
		item.WorkFormat = &f
	}
	c.JSON(http.StatusCreated, item)
}

func (s *Server) UpdateVacancy(c *gin.Context, id openapi_types.UUID) {
	owner, ok := owner(c)
	if !ok {
		return
	}
	var body api.VacancyInput
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "некорректное тело запроса")
		return
	}
	row := s.pool.QueryRow(c.Request.Context(), `
		UPDATE vacancies v
		SET title = coalesce(nullif($3,''), v.title),
		    description = coalesce(nullif($4,''), v.description),
		    salary_min = case when $5 > 0 then $5 else v.salary_min end,
		    salary_max = case when $6 > 0 then $6 else v.salary_max end,
		    updated_at = now()
		FROM companies c
		WHERE v.id = $1 AND v.company_id = c.id AND c.owner_user_id = $2
		RETURNING v.id, v.company_id, v.title, v.description, v.category_id, v.stack,
		          v.salary_min, v.salary_max, v.work_format, v.status, v.published_at, v.created_at`,
		id, owner, body.Title, body.Description, body.SalaryMin, body.SalaryMax)
	item, err := readVacancy(row)
	if err != nil {
		httpx.GinError(c, http.StatusNotFound, "not_found", "вакансия не найдена")
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) DeleteVacancy(c *gin.Context, id openapi_types.UUID) {
	owner, ok := owner(c)
	if !ok {
		return
	}
	tag, err := s.pool.Exec(c.Request.Context(), `
		DELETE FROM vacancies v USING companies c
		WHERE v.id = $1 AND v.company_id = c.id AND c.owner_user_id = $2`, id, owner)
	if err != nil || tag.RowsAffected() == 0 {
		httpx.GinError(c, http.StatusNotFound, "not_found", "вакансия не найдена")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) PublishVacancy(c *gin.Context, id openapi_types.UUID) {
	owner, ok := owner(c)
	if !ok {
		return
	}
	row := s.pool.QueryRow(c.Request.Context(), `
		UPDATE vacancies v
		SET status = 'published', published_at = now(), updated_at = now()
		FROM companies c
		WHERE v.id = $1 AND v.company_id = c.id AND c.owner_user_id = $2
		RETURNING v.id, v.company_id, v.title, v.description, v.category_id, v.stack,
		          v.salary_min, v.salary_max, v.work_format, v.status, v.published_at, v.created_at`,
		id, owner)
	item, err := readVacancy(row)
	if err != nil {
		httpx.GinError(c, http.StatusNotFound, "not_found", "вакансия не найдена")
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) ListPublicVacancies(c *gin.Context, _ api.ListPublicVacanciesParams) {
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT id, company_id, title, description, category_id, stack,
		       salary_min, salary_max, work_format, status, published_at, created_at
		FROM vacancies
		WHERE status IN ('published', 'active')
		ORDER BY created_at DESC`)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "ошибка чтения вакансий")
		return
	}
	defer rows.Close()
	out := []api.Vacancy{}
	for rows.Next() {
		item, err := readVacancy(rows)
		if err != nil {
			httpx.GinError(c, http.StatusInternalServerError, "internal", "ошибка чтения вакансии")
			return
		}
		out = append(out, item)
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) GetPublicVacancy(c *gin.Context, id openapi_types.UUID) {
	row := s.pool.QueryRow(c.Request.Context(), `
		SELECT id, company_id, title, description, category_id, stack,
		       salary_min, salary_max, work_format, status, published_at, created_at
		FROM vacancies
		WHERE id = $1`, id)
	item, err := readVacancy(row)
	if err != nil {
		httpx.GinError(c, http.StatusNotFound, "not_found", "вакансия не найдена")
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) company(c *gin.Context, ownerID uuid.UUID) (api.Company, error) {
	var item api.Company
	err := s.pool.QueryRow(c.Request.Context(), `
		SELECT id, name, description, industry, website, size FROM companies WHERE owner_user_id = $1`, ownerID).
		Scan(&item.Id, &item.Name, &item.Description, &item.Industry, &item.Website, &item.Size)
	return item, err
}

func (s *Server) notReady(c *gin.Context) {
	httpx.GinError(c, http.StatusNotImplemented, "not_implemented", "эта операция появится вместе с публикацией вакансий")
}
func (s *Server) writeMissing(c *gin.Context, _ openapi_types.UUID) { s.notReady(c) }

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
	r.GET("/vacancies", w.ListPublicVacancies)
	r.GET("/vacancies/:id", w.GetPublicVacancy)
	auth := r.Group("")
	auth.Use(httpx.GinJWTAuth(signer), httpx.GinRequireVerifiedEmail(), httpx.GinRequireRole("employer", "admin"))
	auth.GET("/me/company", w.GetMyCompany)
	auth.PUT("/me/company", w.UpsertMyCompany)
	auth.GET("/me/needs", w.ListMyNeeds)
	auth.POST("/me/needs", w.CreateNeed)
	auth.PATCH("/me/needs/:id", w.UpdateNeed)
	auth.DELETE("/me/needs/:id", w.DeleteNeed)
	auth.GET("/me/needs/:id/matches", w.GetNeedMatches)
	auth.GET("/me/vacancies", w.ListMyVacancies)
	auth.POST("/me/vacancies", w.CreateVacancy)
	auth.PATCH("/me/vacancies/:id", w.UpdateVacancy)
	auth.DELETE("/me/vacancies/:id", w.DeleteVacancy)
	auth.POST("/me/vacancies/:id/publish", w.PublishVacancy)
	return r
}

func owner(c *gin.Context) (uuid.UUID, bool) {
	user, ok := httpx.UserFromGin(c)
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
		return uuid.Nil, false
	}
	id, err := uuid.Parse(user.ID)
	if err != nil {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "bad user id")
		return uuid.Nil, false
	}
	return id, true
}

func bindNeed(c *gin.Context) (api.NeedInput, bool) {
	var body api.NeedInput
	if err := c.ShouldBindJSON(&body); err != nil || body.Title == "" || body.Description == "" || body.SalaryMax < body.SalaryMin {
		httpx.GinError(c, http.StatusBadRequest, "invalid_body", "проверьте название, описание и вилку")
		return body, false
	}
	return body, true
}

func readNeed(row pgx.Row) (api.Need, error) {
	var item api.Need
	var stack []uuid.UUID
	var format *string
	var status string
	err := row.Scan(&item.Id, &item.Title, &item.Description, &item.CategoryId, &item.SpecializationId, &item.GradeId, &stack, &item.SalaryMin, &item.SalaryMax, &item.Currency, &format, &status)
	item.Stack = uuidOut(stack)
	item.Status = api.NeedStatus(status)
	if format != nil {
		value := api.NeedWorkFormat(*format)
		item.WorkFormat = &value
	}
	return item, err
}

func uuidOut(in []uuid.UUID) *[]openapi_types.UUID {
	if len(in) == 0 {
		return nil
	}
	out := make([]openapi_types.UUID, len(in))
	for i, id := range in {
		out[i] = id
	}
	return &out
}

func str(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
func uuids(v *[]openapi_types.UUID) []uuid.UUID {
	if v == nil {
		return []uuid.UUID{}
	}
	out := make([]uuid.UUID, len(*v))
	for i, id := range *v {
		out[i] = uuid.UUID(id)
	}
	return out
}
func formatOfNeed(v *api.NeedInputWorkFormat) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}
func statusOf(v *api.NeedInputStatus) string {
	if v == nil || *v == "" {
		return "active"
	}
	return string(*v)
}

func formatOfVacancy(v *api.VacancyInputWorkFormat) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func readVacancy(row pgx.Row) (api.Vacancy, error) {
	var item api.Vacancy
	var id uuid.UUID
	var companyID uuid.UUID
	var categoryID *uuid.UUID
	var stack []uuid.UUID
	var format *string
	var status string
	var pubAt *time.Time
	var createdAt time.Time
	err := row.Scan(&id, &companyID, &item.Title, &item.Description, &categoryID, &stack, &item.SalaryMin, &item.SalaryMax, &format, &status, &pubAt, &createdAt)
	if err != nil {
		return item, err
	}
	item.Id = &id
	item.CompanyId = &companyID
	item.CategoryId = categoryID
	item.Stack = uuidOut(stack)
	st := api.VacancyStatus(status)
	item.Status = &st
	if format != nil {
		f := api.VacancyWorkFormat(*format)
		item.WorkFormat = &f
	}
	item.PublishedAt = pubAt
	item.CreatedAt = &createdAt
	return item, nil
}
