package application

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
)

type CreateResumeInput struct {
	Title     string
	Content   string
	IsPrimary bool
}

func (s *Service) ListResumes(ctx context.Context, userID string) ([]domain.Resume, error) {
	return s.resumes.ListByUser(ctx, userID)
}

func (s *Service) CreateResume(ctx context.Context, userID string, in CreateResumeInput) (*domain.Resume, error) {
	r := &domain.Resume{
		UserID:    userID,
		Title:     in.Title,
		Source:    domain.ResumeSourceManual,
		Content:   &in.Content,
		IsPrimary: in.IsPrimary,
	}
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		if in.IsPrimary {
			if err := s.resumes.UnsetPrimary(ctx, userID); err != nil {
				return err
			}
		}
		return s.resumes.Create(ctx, r)
	})
	return r, err
}

func (s *Service) UploadResumePDF(ctx context.Context, userID string, file *multipart.FileHeader) (*domain.Resume, error) {
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
		return nil, domain.ErrInvalidFileType
	}
	if file.Size > s.cfg.MaxUploadBytes {
		return nil, domain.ErrFileTooLarge
	}

	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open upload: %w", err)
	}
	defer f.Close()

	path, err := s.files.Save(ctx, userID, file.Filename, f, s.cfg.MaxUploadBytes)
	if err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	// Парсинг отложен — заглушка. В будущем здесь PDF→text→NER.
	r := &domain.Resume{
		UserID:   userID,
		Title:    strings.TrimSuffix(file.Filename, ".pdf"),
		Source:   domain.ResumeSourcePDFUpload,
		FilePath: &path,
		Parsed:   map[string]any{},
	}
	if err := s.resumes.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}
