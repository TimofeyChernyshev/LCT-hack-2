package jwtx

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Signer — интерфейс, чтобы заменить HS256 на RS256/JWKS
// при переходе на Keycloak без переписывания вызовов.
type Signer interface {
	Sign(claims Claims) (string, error)
	Parse(token string) (*Claims, error)
}

type HS256Signer struct {
	secret []byte
	issuer string
}

func NewHS256Signer(secret, issuer string) *HS256Signer {
	return &HS256Signer{secret: []byte(secret), issuer: issuer}
}

func (s *HS256Signer) Sign(c Claims) (string, error) {
	if c.Issuer == "" {
		c.Issuer = s.issuer
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":            c.Subject,
		"iss":            c.Issuer,
		"aud":            c.Audience,
		"exp":            c.ExpiresAt,
		"iat":            c.IssuedAt,
		"nbf":            c.NotBefore,
		"jti":            c.JWTID,
		"role":           c.Role,
		"email":          c.Email,
		"email_verified": c.EmailVerified,
	})
	return t.SignedString(s.secret)
}

func (s *HS256Signer) Parse(tokenStr string) (*Claims, error) {
	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	}, jwt.WithIssuer(s.issuer), jwt.WithExpirationRequired())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	mc, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return nil, ErrTokenInvalid
	}

	c := &Claims{
		Subject:       str(mc["sub"]),
		Issuer:        str(mc["iss"]),
		ExpiresAt:     int64(num(mc["exp"])),
		IssuedAt:      int64(num(mc["iat"])),
		NotBefore:     int64(num(mc["nbf"])),
		JWTID:         str(mc["jti"]),
		Role:          str(mc["role"]),
		Email:         str(mc["email"]),
		EmailVerified: boolv(mc["email_verified"]),
	}
	if aud, ok := mc["aud"].([]any); ok {
		for _, a := range aud {
			c.Audience = append(c.Audience, str(a))
		}
	}
	if err := c.Valid(); err != nil {
		return nil, err
	}
	return c, nil
}

func str(v any) string  { s, _ := v.(string); return s }
func num(v any) float64 { f, _ := v.(float64); return f }
func boolv(v any) bool  { b, _ := v.(bool); return b }

// Helper для удобной генерации claims.
func NewAccessClaims(userID, role, email string, verified bool, issuer string, ttl time.Duration) Claims {
	now := time.Now()
	return Claims{
		Subject:       userID,
		Issuer:        issuer,
		ExpiresAt:     now.Add(ttl).Unix(),
		IssuedAt:      now.Unix(),
		Role:          role,
		Email:         email,
		EmailVerified: verified,
	}
}
