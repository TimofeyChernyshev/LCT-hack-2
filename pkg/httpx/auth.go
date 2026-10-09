package httpx

import (
	"context"
	"net/http"
	"strings"

	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
)

type ctxUserKey struct{}

type UserInfo struct {
	ID            string
	Role          string
	Email         string
	EmailVerified bool
}

// JWTAuth — валидация access-токена. Кладёт UserInfo в контекст.
func JWTAuth(signer jwtx.Signer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				WriteError(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
				return
			}
			tok := strings.TrimPrefix(auth, "Bearer ")
			claims, err := signer.Parse(tok)
			if err != nil {
				WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid token")
				return
			}
			u := UserInfo{
				ID:            claims.Subject,
				Role:          claims.Role,
				Email:         claims.Email,
				EmailVerified: claims.EmailVerified,
			}
			ctx := context.WithValue(r.Context(), ctxUserKey{}, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole — RBAC. Пропускает только указанные роли.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
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

// RequireVerifiedEmail — блокирует действия до подтверждения email.
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

func UserFromContext(ctx context.Context) (UserInfo, bool) {
	u, ok := ctx.Value(ctxUserKey{}).(UserInfo)
	return u, ok
}
