package http

import (
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/application"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/domain"
	apiauth "github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/infrastructure/http/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func toUserResponse(m *application.Me) (apiauth.User, error) {
	id, err := uuidToAPI(m.ID)
	if err != nil {
		return apiauth.User{}, err
	}
	return apiauth.User{
		Id:              id,
		Email:           openapi_types.Email(m.Email),
		Role:            apiauth.UserRole(m.Role),
		Status:          apiauth.UserStatus(m.Status),
		EmailVerifiedAt: m.EmailVerifiedAt,
		CreatedAt:       m.CreatedAt,
	}, nil
}

func toTokenPairResponse(p *application.TokenPair, me *application.Me) (apiauth.TokenPair, error) {
	user, err := toUserResponse(me)
	if err != nil {
		return apiauth.TokenPair{}, err
	}
	return apiauth.TokenPair{
		AccessToken:      p.AccessToken,
		RefreshToken:     p.RefreshToken,
		AccessExpiresAt:  p.AccessExpiresAt,
		RefreshExpiresAt: p.RefreshExpiresAt,
		User:             user,
	}, nil
}

func toConsentsResponse(list []domain.Consent) ([]apiauth.Consent, error) {
	out := make([]apiauth.Consent, 0, len(list))
	for _, c := range list {
		id, err := uuidToAPI(c.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, apiauth.Consent{
			Id:        id,
			Type:      apiauth.ConsentType(c.Type),
			Version:   c.Version,
			GrantedAt: c.GrantedAt.UTC(),
			RevokedAt: c.RevokedAt,
		})
	}
	return out, nil
}
