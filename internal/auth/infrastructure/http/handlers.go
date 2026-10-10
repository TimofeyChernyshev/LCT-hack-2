package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/application"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/domain"
	apiauth "github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Handlers struct {
	svc *application.Service
}

// Проверка на этапе компиляции: реализуем ServerInterface из сгенерированного кода.
var _ apiauth.ServerInterface = (*Handlers)(nil)

func NewHandlers(svc *application.Service) *Handlers {
	return &Handlers{svc: svc}
}

//  Healthz

func (h *Handlers) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

//  Register

func (h *Handlers) Register(c *gin.Context) {
	var req apiauth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}

	// ВНИМАНИЕ: в текущей спеке нет полей согласий.
	// Когда вернёшь их в OpenAPI (RegisterRequest.consentPdn, consentProfilePublication),
	// замени true на req.ConsentPdn и req.ConsentProfilePublication.
	out, err := h.svc.Register(c.Request.Context(), application.RegisterInput{
		Email:                string(req.Email),
		Password:             req.Password,
		Role:                 domain.Role(req.Role),
		ConsentPDN:           req.ConsentPdn,
		ConsentProfilePublic: req.ConsentProfilePublication,
		IP:                   clientIP(c),
		UserAgent:            c.Request.UserAgent(),
	})
	if err != nil {
		writeDomainError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"userId": out.UserID})
}

//  ConfirmEmail

func (h *Handlers) ConfirmEmail(c *gin.Context) {
	var req apiauth.TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	if err := h.svc.ConfirmEmail(c.Request.Context(), req.Token); err != nil {
		writeDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

//  ResendConfirmation

func (h *Handlers) ResendConfirmation(c *gin.Context) {
	var req apiauth.ResendConfirmationJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	if err := h.svc.ResendConfirmation(c.Request.Context(), string(req.Email)); err != nil {
		writeDomainError(c, err)
		return
	}
	c.Status(http.StatusAccepted)
}

//  Login

func (h *Handlers) Login(c *gin.Context) {
	var req apiauth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}

	ctx := c.Request.Context()
	pair, err := h.svc.Login(ctx, application.LoginInput{
		Email:     string(req.Email),
		Password:  req.Password,
		IP:        clientIP(c),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		writeDomainError(c, err)
		return
	}

	// TokenPair в спеке требует полный User, а application.TokenPair
	// несёт только id/role/email. Дочитываем профиль.
	me, err := h.svc.GetMe(ctx, pair.UserID)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}
	resp, err := toTokenPairResponse(pair, me)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}
	c.JSON(http.StatusOK, resp)
}

//  Refresh

func (h *Handlers) Refresh(c *gin.Context) {
	var req apiauth.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}

	ctx := c.Request.Context()
	pair, err := h.svc.Refresh(ctx, req.RefreshToken, clientIP(c), c.Request.UserAgent())
	if err != nil {
		writeDomainError(c, err)
		return
	}
	me, err := h.svc.GetMe(ctx, pair.UserID)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}
	resp, err := toTokenPairResponse(pair, me)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}
	c.JSON(http.StatusOK, resp)
}

//  Logout

func (h *Handlers) Logout(c *gin.Context) {
	var req apiauth.LogoutJSONRequestBody
	_ = c.ShouldBindJSON(&req)
	_ = h.svc.Logout(c.Request.Context(), req.RefreshToken)
	c.Status(http.StatusNoContent)
}

//  ForgotPassword

func (h *Handlers) ForgotPassword(c *gin.Context) {
	var req apiauth.ForgotPasswordJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	// Не палим существование email — всегда 202.
	_ = h.svc.ForgotPassword(c.Request.Context(), string(req.Email))
	c.Status(http.StatusAccepted)
}

//  ResetPassword

func (h *Handlers) ResetPassword(c *gin.Context) {
	var req apiauth.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	if err := h.svc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		writeDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

//  Me

func (h *Handlers) Me(c *gin.Context) {
	u, ok := httpx.UserFromGin(c)
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
		return
	}
	me, err := h.svc.GetMe(c.Request.Context(), u.ID)
	if err != nil {
		writeDomainError(c, err)
		return
	}

	resp, err := toUserResponse(me)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}
	c.JSON(http.StatusOK, resp)
}

//  DeleteMe

func (h *Handlers) DeleteMe(c *gin.Context) {
	u, ok := httpx.UserFromContext(c.Request.Context())
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
		return
	}
	var req apiauth.DeleteMeJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.GinError(c, http.StatusBadRequest, "invalid_json", "invalid body")
		return
	}
	if err := h.svc.DeleteMe(c.Request.Context(), u.ID, req.Password, clientIP(c), c.Request.UserAgent()); err != nil {
		writeDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

//  ExternalCallback (задел под Keycloak, пока не реализован)

func (h *Handlers) ExternalCallback(c *gin.Context, provider string) {
	_ = provider
	httpx.GinError(c, http.StatusNotImplemented, "not_implemented", "external login not implemented yet")
}

//  InternalGetUser (для других сервисов)

func (h *Handlers) InternalGetUser(c *gin.Context, userId openapi_types.UUID) {
	me, err := h.svc.GetMe(c.Request.Context(), userId.String())
	if err != nil {
		writeDomainError(c, err)
		return
	}
	resp, err := toUserResponse(me)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}
	c.JSON(http.StatusOK, resp)
}

// consent

func (h *Handlers) ListConsents(c *gin.Context) {
	ctx := c.Request.Context()

	u, ok := httpx.UserFromGin(c)
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
		return
	}

	list, err := h.svc.ListConsents(ctx, u.ID)
	if err != nil {
		writeDomainError(c, err)
		return
	}

	resp, err := toConsentsResponse(list)
	if err != nil {
		httpx.GinError(c, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handlers) GrantConsent(c *gin.Context, type_ apiauth.GrantConsentParamsType) {
	ctx := c.Request.Context()

	u, ok := httpx.UserFromGin(c)
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
		return
	}

	t := domain.ConsentType(type_)
	if !t.Valid() {
		httpx.GinError(c, http.StatusBadRequest, "invalid_consent_type", "unknown consent type")
		return
	}

	if err := h.svc.GrantConsent(ctx, u.ID, t, clientIP(c), c.Request.UserAgent()); err != nil {
		writeDomainError(c, err)
		return
	}

	slog.InfoContext(ctx, "consent granted", "user_id", u.ID, "type", type_)
	c.Status(http.StatusNoContent)
}

func (h *Handlers) RevokeConsent(c *gin.Context, type_ apiauth.RevokeConsentParamsType) {
	ctx := c.Request.Context()

	u, ok := httpx.UserFromGin(c)
	if !ok {
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", "no user")
		return
	}

	t := domain.ConsentType(type_)
	if !t.Valid() {
		httpx.GinError(c, http.StatusBadRequest, "invalid_consent_type", "unknown consent type")
		return
	}

	if err := h.svc.RevokeConsent(ctx, u.ID, t); err != nil {
		writeDomainError(c, err)
		return
	}

	slog.InfoContext(ctx, "consent revoked", "user_id", u.ID, "type", type_)
	c.Status(http.StatusNoContent)
}

// helpers

func writeDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEmailTaken):
		httpx.GinError(c, http.StatusConflict, "email_taken", err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrInvalidToken):
		httpx.GinError(c, http.StatusUnauthorized, "unauthorized", err.Error())
	case errors.Is(err, domain.ErrEmailNotVerified):
		httpx.GinError(c, http.StatusForbidden, "email_not_verified", err.Error())
	case errors.Is(err, domain.ErrUserBlocked),
		errors.Is(err, domain.ErrUserLocked):
		httpx.GinError(c, http.StatusForbidden, "blocked", err.Error())
	case errors.Is(err, domain.ErrConsentRequired),
		errors.Is(err, domain.ErrInvalidConsentType):
		httpx.GinError(c, http.StatusBadRequest, "consent_required", err.Error())
	case errors.Is(err, domain.ErrWeakPassword),
		errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidRole):
		httpx.GinError(c, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		httpx.GinError(c, http.StatusNotFound, "not_found", err.Error())
	default:
		httpx.GinError(c, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func clientIP(c *gin.Context) string {
	if h := c.GetHeader("X-Forwarded-For"); h != "" {
		return h
	}
	return c.ClientIP()
}

// uuidToAPI — конвертирует строковый UUID в тип, используемый кодогеном.
func uuidToAPI(s string) (openapi_types.UUID, error) {
	return uuid.Parse(s)
}
