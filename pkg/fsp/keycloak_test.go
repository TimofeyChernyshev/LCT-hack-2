package fsp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeycloakClaimsHelper_ExtractFSPID(t *testing.T) {
	helper := NewKeycloakClaimsHelper()

	// 1. Direct fsp_id
	claims1 := map[string]any{"fsp_id": "FSP-RU-77-00101"}
	id1, ok1 := helper.ExtractFSPID(claims1)
	assert.True(t, ok1)
	assert.Equal(t, "FSP-RU-77-00101", id1)

	// 2. Direct fsp_member_id
	claims2 := map[string]any{"fsp_member_id": "FSP-10002"}
	id2, ok2 := helper.ExtractFSPID(claims2)
	assert.True(t, ok2)
	assert.Equal(t, "FSP-10002", id2)

	// 3. Nested attributes
	claims3 := map[string]any{
		"attributes": map[string]any{
			"fsp_id": []any{"FSP-RU-16-00303"},
		},
	}
	id3, ok3 := helper.ExtractFSPID(claims3)
	assert.True(t, ok3)
	assert.Equal(t, "FSP-RU-16-00303", id3)

	// 4. Federated identity
	claims4 := map[string]any{
		"federated_identities": []any{
			map[string]any{
				"identity_provider": "fsp-realm",
				"user_id":           "FSP-RU-54-00404",
			},
		},
	}
	id4, ok4 := helper.ExtractFSPID(claims4)
	assert.True(t, ok4)
	assert.Equal(t, "FSP-RU-54-00404", id4)

	// 5. Missing
	claims5 := map[string]any{"email": "user@example.com"}
	_, ok5 := helper.ExtractFSPID(claims5)
	assert.False(t, ok5)
}

func TestKeycloakClaimsHelper_IsKeycloakToken(t *testing.T) {
	helper := NewKeycloakClaimsHelper()

	assert.True(t, helper.IsKeycloakToken(map[string]any{"iss": "http://keycloak.fsp.local/realms/fsp"}))
	assert.True(t, helper.IsKeycloakToken(map[string]any{"azp": "fsp-frontend"}))
	assert.False(t, helper.IsKeycloakToken(map[string]any{"iss": "other-issuer"}))
}
