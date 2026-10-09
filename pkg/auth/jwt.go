package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	CtxUserIDKey      = "userID"
	CtxUserRoleKey     = "userRole"
	CtxFSPMemberIDKey = "fspMemberID"
	HeaderUserID       = "X-User-ID"
	HeaderUserRole     = "X-User-Role"
	HeaderFSPID        = "X-FSP-ID"
)

var (
	ErrMissingToken = errors.New("missing or invalid authorization header")
	ErrInvalidToken = errors.New("invalid token")
)

type Claims struct {
	UserID      uuid.UUID `json:"sub"`
	Role        string    `json:"role"`
	FSPMemberID string    `json:"fsp_id,omitempty"`
	FSPID       string    `json:"fsp_member_id,omitempty"`
	jwt.RegisteredClaims
}

type JWTValidator struct {
	secret []byte
	issuer string
}

func NewJWTValidator(secret, issuer string) *JWTValidator {
	return &JWTValidator{
		secret: []byte(secret),
		issuer: issuer,
	}
}

func (v *JWTValidator) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if v.issuer != "" && claims.Issuer != "" && claims.Issuer != v.issuer {
		return nil, fmt.Errorf("%w: issuer mismatch", ErrInvalidToken)
	}

	return claims, nil
}

// GenerateTestToken is a utility for testing and dev environments
func (v *JWTValidator) GenerateTestToken(userID uuid.UUID, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    v.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(v.secret)
}

// Middleware creates a Gin middleware that extracts and validates JWT tokens.
// In development mode, it also supports X-User-ID / X-User-Role headers for testing convenience.
func Middleware(validator *JWTValidator, isDev bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				claims, err := validator.ValidateToken(parts[1])
				if err == nil {
					c.Set(CtxUserIDKey, claims.UserID)
					c.Set(CtxUserRoleKey, claims.Role)
					if claims.FSPMemberID != "" {
						c.Set(CtxFSPMemberIDKey, claims.FSPMemberID)
					} else if claims.FSPID != "" {
						c.Set(CtxFSPMemberIDKey, claims.FSPID)
					}
					c.Next()
					return
				}
			}
		}

		// Development fallback via custom headers
		if isDev {
			if devUserID := c.GetHeader(HeaderUserID); devUserID != "" {
				if uid, err := uuid.Parse(devUserID); err == nil {
					role := c.GetHeader(HeaderUserRole)
					if role == "" {
						role = "candidate"
					}
					c.Set(CtxUserIDKey, uid)
					c.Set(CtxUserRoleKey, role)
					if devFSP := c.GetHeader(HeaderFSPID); devFSP != "" {
						c.Set(CtxFSPMemberIDKey, devFSP)
					}
					c.Next()
					return
				}
			}
		}

		// If path doesn't require authentication (e.g., healthcheck, internal routes, or public FSP registry), proceed
		path := c.Request.URL.Path
		if path == "/healthz" || strings.HasPrefix(path, "/internal/") || strings.HasPrefix(path, "/fsp/registry/") {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "Authorization required",
		})
	}
}

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get(CtxUserIDKey)
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	uid, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID in context is not a valid UUID")
	}
	return uid, nil
}

// GetUserRole extracts the user role from the Gin context
func GetUserRole(c *gin.Context) string {
	val, exists := c.Get(CtxUserRoleKey)
	if !exists {
		return ""
	}
	role, _ := val.(string)
	return role
}

// GetFSPMemberID extracts the FSP member ID if present in the context
func GetFSPMemberID(c *gin.Context) string {
	val, exists := c.Get(CtxFSPMemberIDKey)
	if !exists {
		return ""
	}
	fspID, _ := val.(string)
	return fspID
}
