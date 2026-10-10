package http

import (
	"errors"
	"net/http"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/application"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/domain"
	"github.com/gin-gonic/gin"
)

func writeError(c *gin.Context, status int, code, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"code": code, "message": msg})
}

func writeDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrCompanyNotFound),
		errors.Is(err, domain.ErrNeedNotFound),
		errors.Is(err, domain.ErrVacancyNotFound),
		errors.Is(err, application.ErrNotFound):
		writeError(c, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, domain.ErrNotOwner):
		writeError(c, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, domain.ErrSalaryRange):
		writeError(c, http.StatusBadRequest, "salary_range", err.Error())
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(c, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, domain.ErrNotPublished):
		writeError(c, http.StatusNotFound, "not_published", err.Error())
	default:
		writeError(c, http.StatusInternalServerError, "internal_error", "internal error")
	}
}
