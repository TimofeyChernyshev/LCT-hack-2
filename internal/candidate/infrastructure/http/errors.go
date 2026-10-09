package http

import (
	"errors"
	"net/http"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/httpx"
	"github.com/gin-gonic/gin"
)

func writeDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrProfileNotFound),
		errors.Is(err, domain.ErrContactsNotFound),
		errors.Is(err, domain.ErrResumeNotFound),
		errors.Is(err, domain.ErrExperienceNotFound):
		httpx.GinError(c, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, domain.ErrNotOwner):
		httpx.GinError(c, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, domain.ErrSalaryRange):
		httpx.GinError(c, http.StatusBadRequest, "salary_range", err.Error())
	case errors.Is(err, domain.ErrInvalidInput):
		httpx.GinError(c, http.StatusBadRequest, "invalid_input", err.Error())
	default:
		httpx.GinError(c, http.StatusInternalServerError, "internal_error", "internal error")
	}
}
