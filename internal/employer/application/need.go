package application

import (
	"context"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/domain"
)

type NeedInput struct {
	Title              string
	Description        string
	CategoryID         *string
	SpecializationID   *string
	GradeID            *string
	Stack              []string
	SalaryMin          int
	SalaryMax          int
	Currency           *string
	WorkFormat         *string
	Location           *string
	MinExperienceYears *float32
	Status             *string
}

func (s *Service) ListMyNeeds(ctx context.Context, ownerID string) ([]domain.Need, error) {
	c, err := s.companies.GetByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return s.needs.ListByCompany(ctx, c.ID)
}

func (s *Service) CreateNeed(ctx context.Context, ownerID string, in NeedInput) (*domain.Need, error) {
	if in.Title == "" || in.Description == "" {
		return nil, domain.ErrInvalidInput
	}
	if in.SalaryMin > in.SalaryMax {
		return nil, domain.ErrSalaryRange
	}

	var out *domain.Need
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		c, err := s.companies.GetByOwner(ctx, ownerID)
		if err != nil {
			return err
		}
		n := &domain.Need{CompanyID: c.ID}
		applyNeedInput(n, in)
		if err := s.needs.Create(ctx, n); err != nil {
			return err
		}
		out = n
		return nil
	})
	return out, err
}

func (s *Service) UpdateNeed(ctx context.Context, ownerID, needID string, in NeedInput) (*domain.Need, error) {
	if in.SalaryMin > in.SalaryMax {
		return nil, domain.ErrSalaryRange
	}
	var out *domain.Need
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		c, err := s.companies.GetByOwner(ctx, ownerID)
		if err != nil {
			return err
		}
		n, err := s.needs.GetByID(ctx, needID)
		if err != nil {
			return err
		}
		if n.CompanyID != c.ID {
			return domain.ErrNotOwner
		}
		applyNeedInput(n, in)
		if err := s.needs.Update(ctx, n); err != nil {
			return err
		}
		out = n
		return nil
	})
	return out, err
}

func (s *Service) DeleteNeed(ctx context.Context, ownerID, needID string) error {
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		c, err := s.companies.GetByOwner(ctx, ownerID)
		if err != nil {
			return err
		}
		return s.needs.Delete(ctx, needID, c.ID)
	})
}

func applyNeedInput(n *domain.Need, in NeedInput) {
	n.Title = in.Title
	n.Description = in.Description
	n.CategoryID = in.CategoryID
	n.SpecializationID = in.SpecializationID
	n.GradeID = in.GradeID
	if in.Stack != nil {
		n.Stack = in.Stack
	}
	n.SalaryMin = in.SalaryMin
	n.SalaryMax = in.SalaryMax
	if in.Currency != nil {
		n.Currency = *in.Currency
	} else if n.Currency == "" {
		n.Currency = "RUB"
	}
	if in.WorkFormat != nil {
		wf := domain.WorkFormat(*in.WorkFormat)
		n.WorkFormat = &wf
	}
	n.Location = in.Location
	n.MinExperienceYears = in.MinExperienceYears
	if in.Status != nil {
		n.Status = domain.NeedStatus(*in.Status)
	} else if n.Status == "" {
		n.Status = domain.NeedStatusDraft
	}
}
