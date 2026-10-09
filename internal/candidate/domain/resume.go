package domain

import "time"

type ResumeSource string

const (
	ResumeSourceManual    ResumeSource = "manual"
	ResumeSourcePDFUpload ResumeSource = "pdf_upload"
	ResumeSourceGenerated ResumeSource = "generated"
)

type Resume struct {
	ID        string
	UserID    string
	Title     string
	Source    ResumeSource
	Content   *string
	FilePath  *string
	Parsed    map[string]any
	IsPrimary bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
