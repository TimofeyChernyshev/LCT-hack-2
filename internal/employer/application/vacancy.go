package application

import (
	"context"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/domain"
)

type VacancyInput struct {
	Title            string
	Description      string
	CategoryID       *string
	SpecializationID *string
	GradeID          *string
	Stack            []string
	SalaryMin        int
	SalaryMax        int
	Currency         *string
	WorkFormat       *string
	Location         *string
}

func (s *Service) ListMyVacancies(ctx context.Context, ownerID string) ([]domain.Vacancy, error) {
	c, err := s.companies.GetByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return s.vacancies.ListByCompany(ctx, c.ID)
}

func (s *Service) CreateVacancy(ctx context.Context, ownerID string, in VacancyInput) (*domain.Vacancy, error) {
	if in.Title == "" || in.Description == "" {
		return nil, domain.ErrInvalidInput
	}
	if in.SalaryMin > in.SalaryMax {
		return nil, domain.ErrSalaryRange
	}

	var out *domain.Vacancy
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		c, err := s.companies.GetByOwner(ctx, ownerID)
		if err != nil {
			return err
		}
		v := &domain.Vacancy{CompanyID: c.ID}
		applyVacancyInput(v, in)
		if err := s.vacancies.Create(ctx, v); err != nil {
			return err
		}
		out = v
		return nil
	})
	return out, err
}

func (s *Service) UpdateVacancy(ctx context.Context, ownerID, vacancyID string, in VacancyInput) (*domain.Vacancy, error) {
	if in.SalaryMin > in.SalaryMax {
		return nil, domain.ErrSalaryRange
	}
	var out *domain.Vacancy
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		c, err := s.companies.GetByOwner(ctx, ownerID)
		if err != nil {
			return err
		}
		v, err := s.vacancies.GetByID(ctx, vacancyID)
		if err != nil {
			return err
		}
		if v.CompanyID != c.ID {
			return domain.ErrNotOwner
		}
		applyVacancyInput(v, in)
		if err := s.vacancies.Update(ctx, v); err != nil {
			return err
		}
		out = v
		return nil
	})
	return out, err
}

func (s *Service) DeleteVacancy(ctx context.Context, ownerID, vacancyID string) error {
	return s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		c, err := s.companies.GetByOwner(ctx, ownerID)
		if err != nil {
			return err
		}
		return s.vacancies.Delete(ctx, vacancyID, c.ID)
	})
}

func (s *Service) PublishVacancy(ctx context.Context, ownerID, vacancyID string) (*domain.Vacancy, error) {
	var out *domain.Vacancy
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		c, err := s.companies.GetByOwner(ctx, ownerID)
		if err != nil {
			return err
		}
		if err := s.vacancies.Publish(ctx, vacancyID, c.ID); err != nil {
			return err
		}
		v, err := s.vacancies.GetByID(ctx, vacancyID)
		if err != nil {
			return err
		}
		out = v
		return nil
	})
	return out, err
}

func (s *Service) ListPublicVacancies(ctx context.Context, filter PublicVacancyFilter) ([]domain.Vacancy, int, error) {
	return s.vacancies.ListPublic(ctx, filter)
}

func (s *Service) GetPublicVacancy(ctx context.Context, id string) (*domain.Vacancy, error) {
	v, err := s.vacancies.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v.Status != domain.VacancyStatusPublished {
		return nil, domain.ErrNotPublished
	}
	return v, nil
}

func applyVacancyInput(v *domain.Vacancy, in VacancyInput) {
	v.Title = in.Title
	v.Description = in.Description
	v.CategoryID = in.CategoryID
	v.SpecializationID = in.SpecializationID
	v.GradeID = in.GradeID
	if in.Stack != nil {
		v.Stack = in.Stack
	}
	v.SalaryMin = in.SalaryMin
	v.SalaryMax = in.SalaryMax
	if in.Currency != nil {
		v.Currency = *in.Currency
	} else if v.Currency == "" {
		v.Currency = "RUB"
	}
	if in.WorkFormat != nil {
		wf := domain.WorkFormat(*in.WorkFormat)
		v.WorkFormat = &wf
	}
	v.Location = in.Location
}
