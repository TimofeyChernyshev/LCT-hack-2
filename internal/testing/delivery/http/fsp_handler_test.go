package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apitesting "github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/auth"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/fsp"
)

func TestHTTP_FSP_GetMyFSP_NoFSP(t *testing.T) {
	router, userID := setupTestServer()

	// GET /me/fsp for candidate who has not linked FSP
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/me/fsp", nil)
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)

	// DoD: Для кандидатов без ФСП интерфейс и алгоритмы работают штатно и прозрачно.
	assert.Equal(t, http.StatusOK, w.Code)

	var profile apitesting.CandidateFSPProfile
	err := json.Unmarshal(w.Body.Bytes(), &profile)
	require.NoError(t, err)

	assert.Equal(t, userID, profile.UserId)
	assert.False(t, profile.HasFsp)
	assert.Nil(t, profile.FspMemberId)
	assert.Equal(t, float32(0.0), profile.FspScore)
	assert.Equal(t, 0, profile.FspRating)
	assert.Equal(t, 0, profile.AchievementsCount)
	assert.Empty(t, profile.Achievements)
	assert.Equal(t, fsp.NoFSPExplanation, profile.Explanation)
}

func TestHTTP_FSP_LinkFSP_Success(t *testing.T) {
	router, userID := setupTestServer()

	// 1. PUT /me/fsp with valid FSP ID
	body, _ := json.Marshal(apitesting.LinkFSPRequest{
		FspMemberId: "FSP-RU-77-00101",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/me/fsp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var linked apitesting.CandidateFSPProfile
	err := json.Unmarshal(w.Body.Bytes(), &linked)
	require.NoError(t, err)

	assert.True(t, linked.HasFsp)
	require.NotNil(t, linked.FspMemberId)
	assert.Equal(t, "FSP-RU-77-00101", *linked.FspMemberId)
	require.NotNil(t, linked.FullName)
	assert.Equal(t, "Смирнов Александр Дмитриевич", *linked.FullName)
	require.NotNil(t, linked.SportsRank)
	assert.Equal(t, "Мастер спорта", *linked.SportsRank)
	assert.Equal(t, 2540, linked.FspRating)
	assert.Greater(t, linked.FspScore, float32(80.0))
	assert.Len(t, linked.Achievements, 3)

	// 2. GET /me/fsp confirms state is now enriched
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/me/fsp", nil)
	req2.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	var confirmed apitesting.CandidateFSPProfile
	_ = json.Unmarshal(w2.Body.Bytes(), &confirmed)
	assert.True(t, confirmed.HasFsp)
	assert.Equal(t, "FSP-RU-77-00101", *confirmed.FspMemberId)
}

func TestHTTP_FSP_LinkFSP_NotFound(t *testing.T) {
	router, userID := setupTestServer()

	body, _ := json.Marshal(apitesting.LinkFSPRequest{
		FspMemberId: "UNKNOWN-FSP-ID-404",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/me/fsp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHTTP_FSP_UnlinkFSP(t *testing.T) {
	router, userID := setupTestServer()

	// Link first
	body, _ := json.Marshal(apitesting.LinkFSPRequest{
		FspMemberId: "FSP-10002",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/me/fsp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// DELETE /me/fsp
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodDelete, "/me/fsp", nil)
	req2.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	var unlinked apitesting.CandidateFSPProfile
	_ = json.Unmarshal(w2.Body.Bytes(), &unlinked)

	assert.False(t, unlinked.HasFsp)
	assert.Nil(t, unlinked.FspMemberId)
	assert.Equal(t, float32(0.0), unlinked.FspScore)
	assert.Equal(t, fsp.NoFSPExplanation, unlinked.Explanation)
}

func TestHTTP_FSP_SyncKeycloak(t *testing.T) {
	router, userID := setupTestServer()

	// POST /me/fsp/sync-keycloak with dev header simulating Keycloak claim
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/me/fsp/sync-keycloak", nil)
	req.Header.Set(auth.HeaderUserID, userID.String())
	req.Header.Set(auth.HeaderFSPID, "FSP-RU-78-00202")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var profile apitesting.CandidateFSPProfile
	err := json.Unmarshal(w.Body.Bytes(), &profile)
	require.NoError(t, err)

	assert.True(t, profile.HasFsp)
	assert.Equal(t, "FSP-RU-78-00202", *profile.FspMemberId)
	assert.Equal(t, "КМС", *profile.SportsRank)
}

func TestHTTP_FSP_InternalEndpoint(t *testing.T) {
	router, userID := setupTestServer()

	// Link user first
	body, _ := json.Marshal(apitesting.LinkFSPRequest{
		FspMemberId: "FSP-RU-16-00303",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/me/fsp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// GET /internal/candidates/{userId}/fsp
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/internal/candidates/"+userID.String()+"/fsp", nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	var internalProfile apitesting.CandidateFSPProfile
	_ = json.Unmarshal(w2.Body.Bytes(), &internalProfile)
	assert.True(t, internalProfile.HasFsp)
	assert.Equal(t, "FSP-RU-16-00303", *internalProfile.FspMemberId)
	assert.Equal(t, "1-й спортивный разряд", *internalProfile.SportsRank)
}

func TestHTTP_FSP_RegistryEndpoints(t *testing.T) {
	router, _ := setupTestServer()

	// 1. GET /fsp/registry/members?q=Смирнов
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/fsp/registry/members?q=Смирнов", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var searchRes struct {
		Total   int                           `json:"total"`
		Members []apitesting.FSPRegistryMember `json:"members"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &searchRes)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, searchRes.Total, 1)
	assert.Equal(t, "Смирнов Александр Дмитриевич", searchRes.Members[0].FullName)

	// 2. GET /fsp/registry/members/{fspId}
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/fsp/registry/members/FSP-RU-77-00101", nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	var member apitesting.FSPRegistryMember
	err = json.Unmarshal(w2.Body.Bytes(), &member)
	require.NoError(t, err)
	assert.Equal(t, "FSP-RU-77-00101", member.FspId)
	assert.Equal(t, "Мастер спорта", member.SportsRank)

	// 3. POST /fsp/registry/verify
	verifyBody, _ := json.Marshal(apitesting.FSPVerificationRequest{
		FspId: "FSP-RU-77-00101",
	})
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodPost, "/fsp/registry/verify", bytes.NewReader(verifyBody))
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)
	var verifyRes apitesting.FSPVerificationResult
	err = json.Unmarshal(w3.Body.Bytes(), &verifyRes)
	require.NoError(t, err)
	assert.True(t, verifyRes.IsValid)
	assert.NotNil(t, verifyRes.Member)
}
