package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/TimofeyChernyshev/LCT-hack-2/internal/dict/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/httpx"
)

type Server struct {
	pool *pgxpool.Pool
}

func NewServer(pool *pgxpool.Pool) *Server {
	return &Server{pool: pool}
}

var _ api.ServerInterface = (*Server)(nil)

func (s *Server) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) ListIndustries(c *gin.Context) {
	rows, err := s.pool.Query(c.Request.Context(), `SELECT id, code, name FROM industries ORDER BY name`)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "справочник не прочитался")
		return
	}
	defer rows.Close()
	out := []api.Industry{}
	for rows.Next() {
		var item api.Industry
		if err := rows.Scan(&item.Id, &item.Code, &item.Name); err != nil {
			httpx.GinError(c, http.StatusInternalServerError, "internal", "справочник не прочитался")
			return
		}
		out = append(out, item)
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) ListSpecializations(c *gin.Context, params api.ListSpecializationsParams) {
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT id, industry_id, code, name, description, is_active
		FROM specializations
		WHERE ($1::uuid IS NULL OR industry_id = $1)
		  AND ($2::bool IS NULL OR is_active = $2)
		ORDER BY name`, params.IndustryId, params.ActiveOnly)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "справочник не прочитался")
		return
	}
	defer rows.Close()
	out := []api.Specialization{}
	for rows.Next() {
		var item api.Specialization
		if err := rows.Scan(&item.Id, &item.IndustryId, &item.Code, &item.Name, &item.Description, &item.IsActive); err != nil {
			httpx.GinError(c, http.StatusInternalServerError, "internal", "справочник не прочитался")
			return
		}
		out = append(out, item)
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) ListGrades(c *gin.Context) {
	rows, err := s.pool.Query(c.Request.Context(), `SELECT id, code, name, rank FROM grades ORDER BY rank`)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "справочник не прочитался")
		return
	}
	defer rows.Close()
	out := []api.Grade{}
	for rows.Next() {
		var item api.Grade
		if err := rows.Scan(&item.Id, &item.Code, &item.Name, &item.Rank); err != nil {
			httpx.GinError(c, http.StatusInternalServerError, "internal", "справочник не прочитался")
			return
		}
		out = append(out, item)
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) ListCategories(c *gin.Context, params api.ListCategoriesParams) {
	items, err := s.categories(c.Request.Context(), params.SpecializationId, params.GradeId, nil)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "справочник не прочитался")
		return
	}
	c.JSON(http.StatusOK, items)
}

func (s *Server) GetCategory(c *gin.Context, id openapi_types.UUID) {
	items, err := s.categories(c.Request.Context(), nil, nil, &id)
	if err != nil || len(items) == 0 {
		httpx.GinError(c, http.StatusNotFound, "not_found", "категория не найдена")
		return
	}
	c.JSON(http.StatusOK, items[0])
}

func (s *Server) ListTechnologies(c *gin.Context, params api.ListTechnologiesParams) {
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT id, code, name, category FROM technologies
		WHERE ($1::text IS NULL OR category = $1)
		ORDER BY name`, params.Category)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "справочник не прочитался")
		return
	}
	defer rows.Close()
	out := []api.Technology{}
	for rows.Next() {
		var item api.Technology
		if err := rows.Scan(&item.Id, &item.Code, &item.Name, &item.Category); err != nil {
			httpx.GinError(c, http.StatusInternalServerError, "internal", "справочник не прочитался")
			return
		}
		out = append(out, item)
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) Snapshot(c *gin.Context) {
	c.JSON(http.StatusOK, api.Snapshot{})
}

func (s *Server) categories(ctx context.Context, specID, gradeID, id *openapi_types.UUID) ([]api.Category, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, specialization_id, grade_id, slug, is_active
		FROM categories
		WHERE ($1::uuid IS NULL OR specialization_id = $1)
		  AND ($2::uuid IS NULL OR grade_id = $2)
		  AND ($3::uuid IS NULL OR id = $3)
		ORDER BY slug`, specID, gradeID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []api.Category{}
	for rows.Next() {
		var item api.Category
		if err := rows.Scan(&item.Id, &item.SpecializationId, &item.GradeId, &item.Slug, &item.IsActive); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
