package fsp

import (
	"strings"
)

// KeycloakClaimsHelper extracts FSP federation attributes from standard Keycloak JWT claims
type KeycloakClaimsHelper struct{}

// NewKeycloakClaimsHelper creates a new helper instance
func NewKeycloakClaimsHelper() *KeycloakClaimsHelper {
	return &KeycloakClaimsHelper{}
}

// ExtractFSPID extracts FSP Member ID from various Keycloak claim locations
func (h *KeycloakClaimsHelper) ExtractFSPID(claims map[string]any) (string, bool) {
	if claims == nil {
		return "", false
	}

	// 1. Direct top-level claims: fsp_id, fsp_member_id, fspId
	directKeys := []string{"fsp_id", "fsp_member_id", "fspId", "fspMemberId", "federated_fsp_id"}
	for _, k := range directKeys {
		if val, ok := claims[k]; ok {
			if s, ok := val.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s), true
			}
		}
	}

	// 2. Keycloak user attributes: "attributes": { "fsp_id": ["FSP-RU-77-00101"] }
	if attrs, ok := claims["attributes"].(map[string]any); ok {
		for _, k := range directKeys {
			if val, ok := attrs[k]; ok {
				if s, ok := val.(string); ok && strings.TrimSpace(s) != "" {
					return strings.TrimSpace(s), true
				}
				if slice, ok := val.([]any); ok && len(slice) > 0 {
					if s, ok := slice[0].(string); ok && strings.TrimSpace(s) != "" {
						return strings.TrimSpace(s), true
					}
				}
				if strSlice, ok := val.([]string); ok && len(strSlice) > 0 {
					if strings.TrimSpace(strSlice[0]) != "" {
						return strings.TrimSpace(strSlice[0]), true
					}
				}
			}
		}
	}

	// 3. Keycloak Federated Identity: "federated_identities": [{"identity_provider": "fsp", "user_id": "FSP-101"}]
	if fedIdentities, ok := claims["federated_identities"].([]any); ok {
		for _, fi := range fedIdentities {
			if m, ok := fi.(map[string]any); ok {
				idp, _ := m["identity_provider"].(string)
				if strings.Contains(strings.ToLower(idp), "fsp") {
					if uid, ok := m["user_id"].(string); ok && strings.TrimSpace(uid) != "" {
						return strings.TrimSpace(uid), true
					}
				}
			}
		}
	}

	return "", false
}

// IsKeycloakToken verifies if token claims indicate Keycloak origin
func (h *KeycloakClaimsHelper) IsKeycloakToken(claims map[string]any) bool {
	if claims == nil {
		return false
	}
	if iss, ok := claims["iss"].(string); ok && (strings.Contains(iss, "realms") || strings.Contains(iss, "keycloak")) {
		return true
	}
	if _, ok := claims["azp"].(string); ok {
		return true
	}
	return false
}
