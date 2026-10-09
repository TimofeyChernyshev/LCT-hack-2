package httpx

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ctxUserKey struct{}

type UserInfo struct {
	ID            string
	Role          string
	Email         string
	EmailVerified bool
}

func WithUser(ctx context.Context, u UserInfo) context.Context {
	return context.WithValue(ctx, ctxUserKey{}, u)
}

func UserFromContext(ctx context.Context) (UserInfo, bool) {
	u, ok := ctx.Value(ctxUserKey{}).(UserInfo)
	return u, ok
}

// UserFromGin — удобный шорткат для хендлеров.
func UserFromGin(c *gin.Context) (UserInfo, bool) {
	return UserFromContext(c.Request.Context())
}
