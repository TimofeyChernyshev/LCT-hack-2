package domain

import "errors"

var (
	ErrCompanyNotFound  = errors.New("company not found")
	ErrNeedNotFound     = errors.New("need not found")
	ErrVacancyNotFound  = errors.New("vacancy not found")
	ErrNotOwner         = errors.New("not owner")
	ErrInvalidInput     = errors.New("invalid input")
	ErrSalaryRange      = errors.New("salary min must be <= max")
	ErrAlreadyPublished = errors.New("vacancy already published")
	ErrNotPublished     = errors.New("vacancy not published")
)
