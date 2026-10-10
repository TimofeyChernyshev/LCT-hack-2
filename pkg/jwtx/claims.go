package jwtx

import "time"

// Claims — OIDC-совместимый набор. При переходе на Keycloak
// поля sub/iss/aud/exp/iat/jti уже будут совпадать с тем, что отдаёт Keycloak.
type Claims struct {
	Subject   string   `json:"sub"`
	Issuer    string   `json:"iss"`
	Audience  []string `json:"aud,omitempty"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
	NotBefore int64    `json:"nbf,omitempty"`
	JWTID     string   `json:"jti"`

	// Кастомные claim-ы платформы
	Role          string `json:"role"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func (c Claims) Valid() error {
	now := time.Now().Unix()
	if c.ExpiresAt != 0 && now >= c.ExpiresAt {
		return ErrTokenExpired
	}
	if c.NotBefore != 0 && now < c.NotBefore {
		return ErrTokenNotYetValid
	}
	return nil
}
