package domain

import "errors"

var (
	ErrProfileNotFound    = errors.New("profile not found")
	ErrContactsNotFound   = errors.New("contacts not found")
	ErrResumeNotFound     = errors.New("resume not found")
	ErrExperienceNotFound = errors.New("experience not found")
	ErrNotOwner           = errors.New("not owner")
	ErrInvalidInput       = errors.New("invalid input")
	ErrSalaryRange        = errors.New("salary min must be <= max")
	ErrRevealExists       = errors.New("reveal already exists")
	ErrInvalidFileType    = errors.New("invalid file type")
	ErrFileTooLarge       = errors.New("file too large")
	ErrInvalidTechnology  = errors.New("invalid technology")
	ErrFSPAlreadyLinked   = errors.New("fsp already linked")
)
