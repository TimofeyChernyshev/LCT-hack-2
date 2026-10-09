package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	apitesting "TimofeyChernyshev/LCT-hack-2/internal/testing/infrastructure/http"
	"TimofeyChernyshev/LCT-hack-2/internal/testing/service"
	"TimofeyChernyshev/LCT-hack-2/pkg/auth"
)

type Handler struct {
	svc *service.TestingService
}

func NewHandler(svc *service.TestingService) *Handler {
	return &Handler{svc: svc}
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

	resp := apitesting.SessionResult{
		SessionId:           id,
		Score:               score32,
		AbilityEstimate:     &theta32,
		ResultingGradeId:    res.ResultingGradeID,
		ResultingCategoryId: res.ResultingCategoryID,
		Decision:            apitesting.SessionResultDecision(res.Decision),
	}

	c.JSON(http.StatusOK, resp)
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

	resp := make([]apitesting.PeriodicTask, 0, len(tasks))
	for _, t := range tasks {
		tID := t.ID
		tCreated := t.CreatedAt
		resp = append(resp, apitesting.PeriodicTask{
			Id:         &tID,
			CategoryId: t.CategoryID,
			Title:      t.Title,
			Body:       t.Body,
			CreatedAt:  &tCreated,
		})
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

	var req apitesting.SubmitPeriodicTaskJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid answer body"})
		return
	}

	if err := h.svc.SubmitPeriodicTask(c.Request.Context(), userID, id, req.Answer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit periodic task"})
		return
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
