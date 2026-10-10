package http

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/TimofeyChernyshev/LCT-hack-2/internal/search/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/httpx"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
)

type hit struct {
	api.SearchHit
	DisplayName string `json:"displayName,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

type Server struct{ pool *pgxpool.Pool }

func NewServer(pool *pgxpool.Pool) *Server { return &Server{pool: pool} }

var _ api.ServerInterface = (*Server)(nil)

func (s *Server) Healthz(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }

func (s *Server) SearchCandidates(c *gin.Context, params api.SearchCandidatesParams) {
	if _, ok := httpx.UserFromGin(c); !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
		return
	}
	limit, offset := 20, 0
	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	if params.Offset != nil && *params.Offset > 0 {
		offset = *params.Offset
	}
	stack := []uuid.UUID{}
	if params.Stack != nil {
		for _, id := range *params.Stack {
			if id != uuid.Nil {
				stack = append(stack, uuid.UUID(id))
			}
		}
	}
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT user_id,
		       trim(both ' ' FROM coalesce(first_name,'') || ' ' || coalesce(last_name,'')),
		       current_category_id, current_grade_id, specialization_id, years_experience,
		       fsp_member_id IS NOT NULL
		FROM candidate_profiles
		WHERE ($1::uuid IS NULL OR specialization_id = $1)
		  AND ($2::uuid IS NULL OR current_grade_id = $2)
		  AND ($3::uuid IS NULL OR current_category_id = $3)
		  AND ($4::bool IS NOT TRUE OR fsp_member_id IS NOT NULL)
		  AND (
		    cardinality($7::uuid[]) = 0
		    OR NOT EXISTS (
		      SELECT 1 FROM unnest($7::uuid[]) AS required(technology_id)
		      WHERE NOT EXISTS (
		        SELECT 1 FROM candidate_technologies ct
		        WHERE ct.user_id = candidate_profiles.user_id
		          AND ct.technology_id = required.technology_id
		      )
		    )
		  )
		ORDER BY updated_at DESC
		LIMIT $5 OFFSET $6`, params.SpecializationId, params.GradeId, params.CategoryId, params.HasFsp, limit, offset, stack)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal", "поиск не выполнился")
		return
	}
	defer rows.Close()
	items := []hit{}
	for rows.Next() {
		var item hit
		var category, grade, spec *uuid.UUID
		var years *float32
		var hasFSP bool
		if err := rows.Scan(&item.UserId, &item.DisplayName, &category, &grade, &spec, &years, &hasFSP); err != nil {
			httpx.GinError(c, http.StatusInternalServerError, "internal", "поиск не выполнился")
			return
		}
		item.CategoryId = orNil(category)
		item.GradeId = orNil(grade)
		item.SpecializationId = orNil(spec)
		item.YearsExperience = years
		item.Score = 1
		item.Reasons = []string{"анкета"}
		item.Explanation = "Кандидат в выдаче по заполненной анкете."
		if hasFSP {
			item.Explanation = "Кандидат в выдаче по анкете, есть привязка ФСП."
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset, "total": len(items)})
}

func (s *Server) ExplainCandidate(c *gin.Context, userID openapi_types.UUID) {
	c.JSON(http.StatusOK, gin.H{"userId": userID, "finalScore": 1, "explanation": "Анкета заполнена.", "factors": []any{}})
}

func (s *Server) CategoryRanking(c *gin.Context, categoryID openapi_types.UUID, params api.CategoryRankingParams) {
	s.SearchCandidates(c, api.SearchCandidatesParams{CategoryId: &categoryID, Limit: params.Limit})
}

func NewRouter(si api.ServerInterface, signer jwtx.Signer, origins []string, env string) *gin.Engine {
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(cors.New(cors.Config{AllowOrigins: origins, AllowMethods: []string{"GET", "OPTIONS"}, AllowHeaders: []string{"Authorization", "Content-Type"}, MaxAge: 5 * time.Minute}))
	w := &api.ServerInterfaceWrapper{Handler: si, ErrorHandler: func(c *gin.Context, err error, code int) {
		c.AbortWithStatusJSON(code, gin.H{"code": "bad_request", "message": err.Error()})
	}}
	r.GET("/healthz", w.Healthz)
	auth := r.Group("")
	auth.Use(httpx.GinJWTAuth(signer), httpx.GinRequireVerifiedEmail(), httpx.GinRequireRole("employer", "admin"))
	auth.GET("/candidates", w.SearchCandidates)
	auth.GET("/candidates/:userId/explain", w.ExplainCandidate)
	auth.GET("/categories/:categoryId/ranking", w.CategoryRanking)
	return r
}

func orNil(id *uuid.UUID) openapi_types.UUID {
	if id == nil {
		return uuid.Nil
	}
	return *id
}
