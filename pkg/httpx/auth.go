package httpx

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
)

var (
	errMissingToken = errors.New("missing bearer token")
	errInvalidToken = errors.New("invalid token")
)

// net/http

func JWTAuth(signer jwtx.Signer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, err := parseUser(signer, r.Header.Get("Authorization"))
			if err != nil {
				WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error())
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), u)))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := roleSet(roles)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFromContext(r.Context())
			if !ok {
				WriteError(w, http.StatusUnauthorized, "unauthorized", "no user")
				return
			}
			if _, ok := allowed[u.Role]; !ok {
				WriteError(w, http.StatusForbidden, "forbidden", "insufficient role")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireVerifiedEmail(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || !u.EmailVerified {
			WriteError(w, http.StatusForbidden, "email_not_verified", "email not verified")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// gin

func GinJWTAuth(signer jwtx.Signer) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, err := parseUser(signer, c.GetHeader("Authorization"))
		if err != nil {
			GinError(c, http.StatusUnauthorized, "unauthorized", err.Error())
			return
		}
		ctx := WithUser(c.Request.Context(), u)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GinRequireRole(roles ...string) gin.HandlerFunc {
	allowed := roleSet(roles)
	return func(c *gin.Context) {
		u, ok := UserFromGin(c)
		if !ok {
			GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
			return
		}
		if _, ok := allowed[u.Role]; !ok {
			GinError(c, http.StatusForbidden, "forbidden", "insufficient role")
			return
		}
		c.Next()
	}
}

func GinRequireVerifiedEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := UserFromGin(c)
		if !ok || !u.EmailVerified {
			GinError(c, http.StatusForbidden, "email_not_verified", "email not verified")
			return
		}
		c.Next()
	}
}

// shared

func parseUser(signer jwtx.Signer, authHeader string) (UserInfo, error) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return UserInfo{}, errMissingToken
	}
	claims, err := signer.Parse(strings.TrimPrefix(authHeader, "Bearer "))
	if err != nil {
		return UserInfo{}, errInvalidToken
	}
	return UserInfo{
		ID:            claims.Subject,
		Role:          claims.Role,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
	}, nil
}

func roleSet(roles []string) map[string]struct{} {
	m := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		m[r] = struct{}{}
	}
	return m
}
