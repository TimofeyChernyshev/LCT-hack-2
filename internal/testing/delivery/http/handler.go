package http

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	apitesting "github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/service"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/auth"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/fsp"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/resume"
)

type TaskAnswerItem struct {
	TaskID      uuid.UUID `json:"taskId"`
	TaskTitle   string    `json:"taskTitle"`
	UserID      uuid.UUID `json:"userId"`
	Email       string    `json:"email"`
	Answer      string    `json:"answer"`
	Kind        string    `json:"kind"`
	SubmittedAt time.Time `json:"submittedAt"`
	Reaction    string    `json:"reaction,omitempty"`
}

type Handler struct {
	svc       *service.TestingService
	answersMu sync.RWMutex
	answers   []TaskAnswerItem
}

func NewHandler(svc *service.TestingService) *Handler {
	return &Handler{
		svc:     svc,
		answers: make([]TaskAnswerItem, 0),
	}
}

// Healthz (GET /healthz)
func (h *Handler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// StartSession (POST /me/sessions)
func (h *Handler) StartSession(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req apitesting.StartSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	sess, err := h.svc.StartSession(c.Request.Context(), userID, req.TargetCategoryId)
	if err != nil {
		if errors.Is(err, service.ErrCooldownActive) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start session", "details": err.Error()})
		return
	}

	resp := apitesting.Session{
		Id:               sess.ID,
		TargetCategoryId: sess.TargetCategoryID,
		Status:           apitesting.SessionStatus(sess.Status),
		StartedAt:        sess.StartedAt,
		FinishedAt:       sess.FinishedAt,
	}

	c.JSON(http.StatusCreated, resp)
}

// ListMySessions (GET /me/sessions)
func (h *Handler) ListMySessions(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessions, err := h.svc.ListMySessions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sessions"})
		return
	}

	resp := make([]apitesting.Session, 0, len(sessions))
	for _, s := range sessions {
		resp = append(resp, apitesting.Session{
			Id:               s.ID,
			TargetCategoryId: s.TargetCategoryID,
			Status:           apitesting.SessionStatus(s.Status),
			StartedAt:        s.StartedAt,
			FinishedAt:       s.FinishedAt,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// GetSession (GET /me/sessions/{id})
func (h *Handler) GetSession(c *gin.Context, id openapi_types.UUID) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sess, err := h.svc.GetSession(c.Request.Context(), userID, id)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		if errors.Is(err, service.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get session"})
		return
	}

	items := make([]apitesting.SessionItem, 0, len(sess.Items))
	for _, it := range sess.Items {
		var topic string
		var tType apitesting.SessionItemType = apitesting.SingleChoice
		if it.Task != nil {
			topic = it.Task.Topic
			tType = apitesting.SessionItemType(it.Task.Type)
		}
		items = append(items, apitesting.SessionItem{
			Id:       it.ID,
			Position: it.Position,
			Type:     tType,
			Topic:    topic,
			Body:     it.RenderedBody,
			Status:   apitesting.SessionItemStatus(it.Status),
		})
	}

	resp := apitesting.SessionWithItems{
		Id:               sess.ID,
		TargetCategoryId: sess.TargetCategoryID,
		Status:           apitesting.SessionWithItemsStatus(sess.Status),
		StartedAt:        sess.StartedAt,
		FinishedAt:       sess.FinishedAt,
		Items:            &items,
	}

	c.JSON(http.StatusOK, resp)
}

// AnswerItem (POST /me/sessions/{id}/answers)
func (h *Handler) AnswerItem(c *gin.Context, id openapi_types.UUID) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req apitesting.AnswerInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid answer body"})
		return
	}

	skipped := req.Skipped != nil && *req.Skipped

	err = h.svc.AnswerItem(c.Request.Context(), userID, id, req.ItemId, req.Answer, skipped)
	if err != nil {
		if errors.Is(err, service.ErrSessionClosed) || errors.Is(err, service.ErrItemAlreadyAnswered) {
			c.JSON(http.StatusConflict, gin.H{"error": "item already answered or session closed"})
			return
		}
		if errors.Is(err, service.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to answer item", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// SubmitSession (POST /me/sessions/{id}/submit)
func (h *Handler) SubmitSession(c *gin.Context, id openapi_types.UUID) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	res, err := h.svc.SubmitSession(c.Request.Context(), userID, id)
	if err != nil {
		if errors.Is(err, service.ErrSessionClosed) {
			c.JSON(http.StatusConflict, gin.H{"error": "session already submitted or expired"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit session", "details": err.Error()})
		return
	}

	theta32 := float32(res.AbilityEstimate)
	score32 := float32(res.Score)

	resp := gin.H{
		"sessionId":           id,
		"score":               score32,
		"abilityEstimate":     theta32,
		"resultingGradeId":    res.ResultingGradeID,
		"resultingCategoryId": res.ResultingCategoryID,
		"decision":            res.Decision,
		"explanation":         res.Explanation,
	}

	c.JSON(http.StatusOK, resp)
}

// SubmitQuestionnaire handles onboarding questionnaire (POST /me/questionnaire)
func (h *Handler) SubmitQuestionnaire(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req domain.QuestionnaireInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid questionnaire body", "details": err.Error()})
		return
	}

	res, err := h.svc.EvaluateQuestionnaire(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to evaluate questionnaire", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetQuestionnaireState handles querying onboarding/cooldown state (GET /me/questionnaire)
func (h *Handler) GetQuestionnaireState(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	state, canChangeAt, canStart, err := h.svc.GetQuestionnaireState(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get questionnaire state", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state":        state,
		"canChangeAt":  canChangeAt,
		"canStartTest": canStart,
	})
}

// ListGradeChanges (GET /me/category-changes)
func (h *Handler) ListGradeChanges(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	canChangeAt, changes, err := h.svc.ListGradeChanges(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list changes"})
		return
	}

	respChanges := make([]apitesting.GradeChange, 0, len(changes))
	for _, ch := range changes {
		chTime := ch.ChangedAt
		r := ch.Reason
		respChanges = append(respChanges, apitesting.GradeChange{
			FromGradeId: ch.FromGradeID,
			ToGradeId:   &ch.ToGradeID,
			Reason:      &r,
			ChangedAt:   &chTime,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"canChangeAt": canChangeAt,
		"changes":     respChanges,
	})
}

// InternalGetCategory (GET /internal/users/{userId}/category)
func (h *Handler) InternalGetCategory(c *gin.Context, userId openapi_types.UUID) {
	catID, gradeID, score, err := h.svc.GetCandidateCategory(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query category"})
		return
	}

	if catID == nil || gradeID == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found for user"})
		return
	}

	testScore := 0.0
	if score != nil {
		testScore = *score
	}

	c.JSON(http.StatusOK, gin.H{
		"categoryId": *catID,
		"gradeId":    *gradeID,
		"testScore":  testScore,
	})
}

// ListPeriodicTasks (GET /me/periodic-tasks)
func (h *Handler) ListPeriodicTasks(c *gin.Context) {
	var categoryID *uuid.UUID
	if catStr := c.Query("categoryId"); catStr != "" {
		if uid, err := uuid.Parse(catStr); err == nil {
			categoryID = &uid
		}
	}

	tasks, err := h.svc.ListPeriodicTasks(c.Request.Context(), categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list periodic tasks"})
		return
	}

	userID, _ := auth.GetUserID(c)

	h.answersMu.RLock()
	userAnswers := make(map[uuid.UUID]TaskAnswerItem)
	if userID != uuid.Nil {
		for _, ans := range h.answers {
			if ans.UserID == userID {
				userAnswers[ans.TaskID] = ans
			}
		}
	}
	h.answersMu.RUnlock()

	resp := make([]gin.H, 0, len(tasks))
	for _, t := range tasks {
		tID := t.ID
		tCreated := t.CreatedAt
		item := gin.H{
			"id":         tID,
			"categoryId": t.CategoryID,
			"title":      t.Title,
			"body":       t.Body,
			"createdAt":  tCreated,
			"status":     "open",
		}
		if ans, ok := userAnswers[tID]; ok {
			item["status"] = "sent"
			item["answer"] = ans.Answer
			item["kind"] = ans.Kind
			item["submittedAt"] = ans.SubmittedAt
			item["reaction"] = ans.Reaction
		}
		resp = append(resp, item)
	}

	c.JSON(http.StatusOK, resp)
}

// SubmitPeriodicTask (POST /me/periodic-tasks/{id}/submit)
func (h *Handler) SubmitPeriodicTask(c *gin.Context, id openapi_types.UUID) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Answer string `json:"answer"`
		Kind   string `json:"kind"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid answer body"})
		return
	}
	if req.Kind == "" {
		req.Kind = "solution"
	}

	if err := h.svc.SubmitPeriodicTask(c.Request.Context(), userID, id, req.Answer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit periodic task"})
		return
	}

	// Capture task title
	taskTitle := "Мини-задача"
	if allTasks, err := h.svc.ListPeriodicTasks(c.Request.Context(), nil); err == nil {
		for _, t := range allTasks {
			if t.ID == id {
				taskTitle = t.Title
				break
			}
		}
	}

	email := "candidate@fsp.local"
	if hEmail := c.GetHeader("X-User-Email"); hEmail != "" {
		email = hEmail
	} else if emailVal, ok := c.Get("userEmail"); ok {
		if emailStr, ok := emailVal.(string); ok && emailStr != "" {
			email = emailStr
		}
	}

	h.answersMu.Lock()
	found := false
	for i := range h.answers {
		if h.answers[i].TaskID == id && h.answers[i].UserID == userID {
			h.answers[i].Answer = req.Answer
			h.answers[i].Kind = req.Kind
			h.answers[i].SubmittedAt = time.Now()
			found = true
			break
		}
	}
	if !found {
		h.answers = append(h.answers, TaskAnswerItem{
			TaskID:      id,
			TaskTitle:   taskTitle,
			UserID:      userID,
			Email:       email,
			Answer:      req.Answer,
			Kind:        req.Kind,
			SubmittedAt: time.Now(),
		})
	}
	h.answersMu.Unlock()

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ListTaskResponses (GET /periodic-tasks)
func (h *Handler) ListTaskResponses(c *gin.Context) {
	h.answersMu.RLock()
	defer h.answersMu.RUnlock()
	c.JSON(http.StatusOK, h.answers)
}

// ReactToAnswer (POST /periodic-tasks/reactions)
func (h *Handler) ReactToAnswer(c *gin.Context) {
	var body struct {
		TaskID   uuid.UUID `json:"taskId"`
		UserID   uuid.UUID `json:"userId"`
		Reaction string    `json:"reaction"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reaction body"})
		return
	}
	h.answersMu.Lock()
	defer h.answersMu.Unlock()
	for i := range h.answers {
		if h.answers[i].TaskID == body.TaskID && h.answers[i].UserID == body.UserID {
			h.answers[i].Reaction = body.Reaction
			break
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// CreatePeriodicTask (POST /periodic-tasks)
func (h *Handler) CreatePeriodicTask(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req apitesting.CreatePeriodicTaskJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid periodic task body"})
		return
	}

	task, err := h.svc.CreatePeriodicTask(c.Request.Context(), userID, req.CategoryId, req.Title, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create periodic task"})
		return
	}

	tID := task.ID
	tCreated := task.CreatedAt
	resp := apitesting.PeriodicTask{
		Id:         &tID,
		CategoryId: task.CategoryID,
		Title:      task.Title,
		Body:       task.Body,
		CreatedAt:  &tCreated,
	}

	c.JSON(http.StatusCreated, resp)
}

// GetMyFSP (GET /me/fsp)
func (h *Handler) GetMyFSP(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	profile, err := h.svc.GetCandidateFSP(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get fsp profile", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapCandidateFSPProfileToAPI(profile))
}

// LinkFSP (PUT /me/fsp)
func (h *Handler) LinkFSP(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req apitesting.LinkFSPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	profile, err := h.svc.LinkFSP(c.Request.Context(), userID, req.FspMemberId, "manual")
	if err != nil {
		if errors.Is(err, service.ErrFSPMemberNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "участник с указанным ID не найден в реестре ФСП"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link fsp profile", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapCandidateFSPProfileToAPI(profile))
}

// UnlinkFSP (DELETE /me/fsp)
func (h *Handler) UnlinkFSP(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	profile, err := h.svc.UnlinkFSP(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlink fsp profile", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapCandidateFSPProfileToAPI(profile))
}

// SyncKeycloakFSP (POST /me/fsp/sync-keycloak)
func (h *Handler) SyncKeycloakFSP(c *gin.Context) {
	userID, err := auth.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 1. Check if token or dev header already set FSP member ID in context
	fspID := auth.GetFSPMemberID(c)
	if fspID != "" {
		profile, err := h.svc.LinkFSP(c.Request.Context(), userID, fspID, "keycloak")
		if err == nil {
			c.JSON(http.StatusOK, mapCandidateFSPProfileToAPI(profile))
			return
		}
	}

	// 2. Alternatively check JSON body for Keycloak claims or raw ID if passed
	var bodyClaims map[string]any
	_ = c.ShouldBindJSON(&bodyClaims)

	profile, err := h.svc.SyncKeycloakFSP(c.Request.Context(), userID, bodyClaims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync keycloak fsp profile", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapCandidateFSPProfileToAPI(profile))
}

// InternalGetCandidateFSP (GET /internal/candidates/{userId}/fsp)
func (h *Handler) InternalGetCandidateFSP(c *gin.Context, userId openapi_types.UUID) {
	profile, err := h.svc.GetCandidateFSP(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get internal candidate fsp", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapCandidateFSPProfileToAPI(profile))
}

// SearchFSPRegistryMembers (GET /fsp/registry/members)
func (h *Handler) SearchFSPRegistryMembers(c *gin.Context, params apitesting.SearchFSPRegistryMembersParams) {
	var query string
	if params.Q != nil {
		query = *params.Q
	}
	var rank fsp.SportsRank
	if params.Rank != nil {
		rank = fsp.SportsRank(*params.Rank)
	}
	var region string
	if params.Region != nil {
		region = *params.Region
	}
	limit := 20
	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	offset := 0
	if params.Offset != nil && *params.Offset >= 0 {
		offset = *params.Offset
	}

	members, total, err := h.svc.SearchFSPRegistry(c.Request.Context(), query, rank, region, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search fsp registry", "details": err.Error()})
		return
	}

	apiMembers := make([]apitesting.FSPRegistryMember, 0, len(members))
	for _, m := range members {
		apiMembers = append(apiMembers, mapFSPRegistryMemberToAPI(m))
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   total,
		"members": apiMembers,
	})
}

// GetFSPRegistryMember (GET /fsp/registry/members/{fspId})
func (h *Handler) GetFSPRegistryMember(c *gin.Context, fspId string) {
	m, err := h.svc.GetFSPRegistryMember(c.Request.Context(), fspId)
	if err != nil {
		if errors.Is(err, fsp.ErrMemberNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "участник не найден в реестре ФСП"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get fsp registry member", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapFSPRegistryMemberToAPI(*m))
}

// VerifyFSPMember (POST /fsp/registry/verify)
func (h *Handler) VerifyFSPMember(c *gin.Context) {
	var req apitesting.FSPVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid verification request body", "details": err.Error()})
		return
	}

	var fullName, cert string
	if req.FullName != nil {
		fullName = *req.FullName
	}
	if req.Certificate != nil {
		cert = *req.Certificate
	}

	res, err := h.svc.VerifyFSPMember(c.Request.Context(), fsp.VerificationRequest{
		FSPID:       req.FspId,
		FullName:    fullName,
		Certificate: cert,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify member", "details": err.Error()})
		return
	}

	apiRes := apitesting.FSPVerificationResult{
		IsValid:    res.IsValid,
		Message:    res.Message,
		VerifiedAt: res.VerifiedAt,
	}
	if res.Member != nil {
		m := mapFSPRegistryMemberToAPI(*res.Member)
		apiRes.Member = &m
	}

	c.JSON(http.StatusOK, apiRes)
}

func mapCandidateFSPProfileToAPI(p *domain.CandidateFSPProfile) apitesting.CandidateFSPProfile {
	achievements := make([]apitesting.FSPAchievement, 0, len(p.Achievements))
	for _, a := range p.Achievements {
		ach := apitesting.FSPAchievement{
			Category:  string(a.Category),
			EventName: a.EventName,
			Weight:    a.Weight,
		}
		if a.ID != "" {
			aID := a.ID
			ach.Id = &aID
		}
		if a.ExternalID != "" {
			extID := a.ExternalID
			ach.ExternalId = &extID
		}
		if a.EventDate != "" {
			ed := a.EventDate
			ach.EventDate = &ed
		}
		if a.Place != nil {
			ach.Place = a.Place
		}
		if a.Score != nil {
			sc := float32(*a.Score)
			ach.Score = &sc
		}
		if a.Badge != "" {
			bd := a.Badge
			ach.Badge = &bd
		}
		if a.Description != "" {
			ds := a.Description
			ach.Description = &ds
		}
		achievements = append(achievements, ach)
	}

	return apitesting.CandidateFSPProfile{
		UserId:             p.UserID,
		HasFsp:             p.HasFSP,
		FspMemberId:        p.FSPMemberID,
		FullName:           p.FullName,
		SportsRank:         p.SportsRank,
		FspRating:          p.FSPRating,
		Region:             p.Region,
		Discipline:         p.Discipline,
		FspScore:           float32(p.FSPScore),
		FspWeightSum:       p.FSPWeightSum,
		AchievementsCount:  p.AchievementsCount,
		BestPlace:          p.BestPlace,
		VerificationSource: p.VerificationSource,
		LinkedAt:           p.LinkedAt,
		Explanation:        p.Explanation,
		Achievements:       achievements,
	}
}

func mapFSPRegistryMemberToAPI(m fsp.Member) apitesting.FSPRegistryMember {
	achievements := make([]apitesting.FSPAchievement, 0, len(m.Achievements))
	for _, a := range m.Achievements {
		ach := apitesting.FSPAchievement{
			Category:  string(a.Category),
			EventName: a.EventName,
			Weight:    a.Weight,
		}
		if a.ID != "" {
			aID := a.ID
			ach.Id = &aID
		}
		if a.ExternalID != "" {
			extID := a.ExternalID
			ach.ExternalId = &extID
		}
		if a.EventDate != "" {
			ed := a.EventDate
			ach.EventDate = &ed
		}
		if a.Place != nil {
			ach.Place = a.Place
		}
		if a.Score != nil {
			sc := float32(*a.Score)
			ach.Score = &sc
		}
		if a.Badge != "" {
			bd := a.Badge
			ach.Badge = &bd
		}
		if a.Description != "" {
			ds := a.Description
			ach.Description = &ds
		}
		achievements = append(achievements, ach)
	}

	status := m.Status
	discipline := string(m.Discipline)

	return apitesting.FSPRegistryMember{
		FspId:        m.FSPID,
		FullName:     m.FullName,
		SportsRank:   string(m.SportsRank),
		Rating:       m.Rating,
		Region:       m.Region,
		Discipline:   &discipline,
		Status:       &status,
		Verified:     m.Verified,
		Achievements: &achievements,
	}
}

// ParseResumePDF (POST /resumes/parse-pdf or POST /me/resumes/upload)
func (h *Handler) ParseResumePDF(c *gin.Context) {
	var fileBytes []byte

	// 1. Multipart file upload ("file" or "resume")
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		file, _, err = c.Request.FormFile("resume")
	}
	if err == nil {
		defer file.Close()
		data, readErr := io.ReadAll(file)
		if readErr == nil && len(data) > 0 {
			fileBytes = data
		}
	}

	// 2. JSON payload with text or base64
	if len(fileBytes) == 0 {
		var jsonReq struct {
			PDFBase64 string `json:"pdf_base64"`
			Text      string `json:"text"`
			Content   string `json:"content"`
		}
		if err := c.ShouldBindJSON(&jsonReq); err == nil {
			if jsonReq.PDFBase64 != "" {
				if dec, decErr := base64.StdEncoding.DecodeString(jsonReq.PDFBase64); decErr == nil {
					fileBytes = dec
				}
			} else if jsonReq.Text != "" {
				fileBytes = []byte(jsonReq.Text)
			} else if jsonReq.Content != "" {
				fileBytes = []byte(jsonReq.Content)
			}
		}
	}

	// 3. Fallback raw body
	if len(fileBytes) == 0 && c.Request.Body != nil {
		raw, _ := io.ReadAll(c.Request.Body)
		if len(raw) > 0 {
			fileBytes = raw
		}
	}

	if len(fileBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "empty resume payload: provide 'file' via multipart form, raw PDF body, or JSON with 'text'/'pdf_base64'",
		})
		return
	}

	parsed, err := resume.ExtractAndParse(fileBytes)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "failed to extract/parse resume PDF: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, parsed)
}

// ExportCandidatePDF renders a standardized branded FSP PDF candidate profile
func (h *Handler) ExportCandidatePDF(c *gin.Context) {
	var profile resume.CandidateExportProfile

	if c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&profile)
	}

	if profile.FullName == "" {
		profile.FullName = c.DefaultQuery("fullName", "Александр Дмитриевич Смирнов")
		profile.Headline = c.DefaultQuery("headline", "Senior Backend Developer")
		profile.SpecializationName = c.DefaultQuery("specialization", "Backend")
		profile.GradeName = c.DefaultQuery("grade", "Senior")
		profile.Location = c.DefaultQuery("location", "Москва")
		profile.TestScore = 95.0
		profile.TestPercentile = 96.0
		profile.HasFSP = true
		profile.SportsRank = "Мастер спорта"
		profile.Skills = []string{"Go", "PostgreSQL", "Redis", "Docker", "Kubernetes", "Kafka"}
		profile.YearsExperience = 6.0
		profile.MaskContacts = true
	}
	if profile.UserID == uuid.Nil {
		profile.UserID = uuid.New()
	}
	if profile.GeneratedAt.IsZero() {
		profile.GeneratedAt = time.Now()
	}

	pdfBytes, err := resume.GenerateCandidatePDF(profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate candidate PDF: " + err.Error()})
		return
	}

	filename := fmt.Sprintf("fsp_profile_%s.pdf", profile.UserID.String()[:8])
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
