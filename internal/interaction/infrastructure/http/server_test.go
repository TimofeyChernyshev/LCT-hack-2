package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/httpx"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
)

func setupTestRouter(server *Server, user httpx.UserInfo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Middleware injects test user into context
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(httpx.WithUser(c.Request.Context(), user))
		c.Next()
	})

	r.POST("/me/invitations", server.CreateInvitation)
	r.GET("/me/invitations", server.ListMyInvitationsAsEmployer)
	r.GET("/me/invitations/incoming", server.ListIncomingInvitations)
	r.POST("/me/invitations/:id/accept", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		server.AcceptInvitation(c, id)
	})
	r.POST("/me/invitations/:id/reject", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		server.RejectInvitation(c, id)
	})
	r.POST("/me/invitations/:id/withdraw", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		server.WithdrawInvitation(c, id)
	})
	r.GET("/me/vacancies/:vacancyId/applications", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("vacancyId"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		server.ListApplicationsForVacancy(c, id)
	})
	r.POST("/me/applications/:id/accept", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		server.AcceptApplication(c, id)
	})
	r.POST("/me/applications/:id/reject", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		server.RejectApplication(c, id)
	})

	return r
}

func TestCreateInvitation_RequiresEmployerRole(t *testing.T) {
	server := NewServer(nil, "", "", "")
	candidateUser := httpx.UserInfo{
		ID:    uuid.New().String(),
		Email: "cand@test.com",
		Role:  "candidate",
	}
	router := setupTestRouter(server, candidateUser)

	body := map[string]any{
		"candidateUserId": uuid.New().String(),
		"message":         "Приглашаем в команду",
		"salaryMin":       200000,
		"salaryMax":       300000,
		"contactChannel":  "telegram",
		"employerContact": "@recruiter",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/me/invitations", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "только работодатель")
}

func TestCreateInvitation_MandatorySalaryFork(t *testing.T) {
	server := NewServer(nil, "", "", "")
	employerUser := httpx.UserInfo{
		ID:    uuid.New().String(),
		Email: "emp@test.com",
		Role:  "employer",
	}
	router := setupTestRouter(server, employerUser)

	tests := []struct {
		name      string
		salaryMin int
		salaryMax int
		wantErr   string
	}{
		{
			name:      "zero salary min and max",
			salaryMin: 0,
			salaryMax: 0,
			wantErr:   "вилка зарплаты в рублях обязательна",
		},
		{
			name:      "negative salary min",
			salaryMin: -100,
			salaryMax: 200000,
			wantErr:   "вилка зарплаты в рублях обязательна",
		},
		{
			name:      "salary min greater than max",
			salaryMin: 350000,
			salaryMax: 250000,
			wantErr:   "вилка зарплаты в рублях обязательна",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]any{
				"candidateUserId": uuid.New().String(),
				"message":         "Приглашаем на позицию Senior Go",
				"salaryMin":       tc.salaryMin,
				"salaryMax":       tc.salaryMax,
				"contactChannel":  "telegram",
				"employerContact": "@hr_lead",
			}
			raw, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/me/invitations", bytes.NewReader(raw))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantErr)
		})
	}
}

func TestCreateInvitation_CurrencyMustBeRUB(t *testing.T) {
	server := NewServer(nil, "", "", "")
	employerUser := httpx.UserInfo{
		ID:    uuid.New().String(),
		Email: "emp@test.com",
		Role:  "employer",
	}
	router := setupTestRouter(server, employerUser)

	usd := "USD"
	body := map[string]any{
		"candidateUserId": uuid.New().String(),
		"message":         "Приглашаем в команду",
		"salaryMin":       200000,
		"salaryMax":       300000,
		"currency":        usd,
		"contactChannel":  "telegram",
		"employerContact": "@hr_lead",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/me/invitations", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "валюта вилки зарплаты должна быть в рублях (RUB)")
}

func TestCreateInvitation_RequiresEmployerContact(t *testing.T) {
	server := NewServer(nil, "", "", "")
	employerUser := httpx.UserInfo{
		ID:    uuid.New().String(),
		Email: "emp@test.com",
		Role:  "employer",
	}
	router := setupTestRouter(server, employerUser)

	body := map[string]any{
		"candidateUserId": uuid.New().String(),
		"message":         "Приглашаем в команду",
		"salaryMin":       200000,
		"salaryMax":       300000,
		"contactChannel":  "telegram",
		"employerContact": "   ",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/me/invitations", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "способ связи с работодателем обязателен")
}

func TestCreateInvitation_RequiresCandidateId(t *testing.T) {
	server := NewServer(nil, "", "", "")
	employerUser := httpx.UserInfo{
		ID:    uuid.New().String(),
		Email: "emp@test.com",
		Role:  "employer",
	}
	router := setupTestRouter(server, employerUser)

	body := map[string]any{
		"candidateUserId": uuid.Nil.String(),
		"message":         "Приглашаем в команду",
		"salaryMin":       200000,
		"salaryMax":       300000,
		"contactChannel":  "telegram",
		"employerContact": "@lead",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/me/invitations", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "укажите кандидата")
}

func TestListIncomingInvitations_RequiresCandidateRole(t *testing.T) {
	server := NewServer(nil, "", "", "")
	employerUser := httpx.UserInfo{
		ID:    uuid.New().String(),
		Email: "emp@test.com",
		Role:  "employer",
	}
	router := setupTestRouter(server, employerUser)

	req := httptest.NewRequest(http.MethodGet, "/me/invitations/incoming", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "только соискатель")
}

func TestListMyInvitationsAsEmployer_RequiresEmployerRole(t *testing.T) {
	server := NewServer(nil, "", "", "")
	candidateUser := httpx.UserInfo{
		ID:    uuid.New().String(),
		Email: "cand@test.com",
		Role:  "candidate",
	}
	router := setupTestRouter(server, candidateUser)

	req := httptest.NewRequest(http.MethodGet, "/me/invitations", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "только работодатель")
}

func TestNewRouter_Registration(t *testing.T) {
	server := NewServer(nil, "", "", "")
	signer := jwtx.NewHS256Signer("test-secret-at-least-32-chars-long!", "fsp-test")
	engine := NewRouter(server, signer, []string{"*"}, "test")
	require.NotNil(t, engine)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListApplicationsForVacancy(t *testing.T) {
	server := NewServer(nil, "", "", "")
	employerUser := httpx.UserInfo{
		ID:    uuid.New().String(),
		Email: "emp@test.com",
		Role:  "employer",
	}
	router := setupTestRouter(server, employerUser)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/me/vacancies/%s/applications", uuid.New()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAcceptApplication_NotFound(t *testing.T) {
	server := NewServer(nil, "", "", "")
	employerUser := httpx.UserInfo{
		ID:    uuid.New().String(),
		Email: "emp@test.com",
		Role:  "employer",
	}
	router := setupTestRouter(server, employerUser)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/me/applications/%s/accept", uuid.New()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRejectApplication_NotFound(t *testing.T) {
	server := NewServer(nil, "", "", "")
	employerUser := httpx.UserInfo{
		ID:    uuid.New().String(),
		Email: "emp@test.com",
		Role:  "employer",
	}
	router := setupTestRouter(server, employerUser)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/me/applications/%s/reject", uuid.New()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
