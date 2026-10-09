package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestJWTValidator_GenerateAndValidate(t *testing.T) {
	secret := "test-secret-at-least-32-chars-long"
	issuer := "fsp-platform"
	val := NewJWTValidator(secret, issuer)

	userID := uuid.New()
	token, err := val.GenerateTestToken(userID, "candidate", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := val.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected user id %s, got %s", userID, claims.UserID)
	}
	if claims.Role != "candidate" {
		t.Errorf("expected role candidate, got %s", claims.Role)
	}
}

func TestJWTMiddleware_DevHeaderFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-at-least-32-chars-long"
	issuer := "fsp-platform"
	val := NewJWTValidator(secret, issuer)

	router := gin.New()
	router.Use(Middleware(val, true))
	router.GET("/protected", func(c *gin.Context) {
		uid, err := GetUserID(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userID": uid.String(), "role": GetUserRole(c)})
	})

	testUID := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(HeaderUserID, testUID.String())
	req.Header.Set(HeaderUserRole, "candidate")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 via dev header, got %d", w.Code)
	}
}
