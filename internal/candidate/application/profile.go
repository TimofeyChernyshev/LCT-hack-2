package application

import (
	"context"
	"errors"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
)

type UpdateProfileInput struct {
	FirstName       *string
	LastName        *string
	MiddleName      *string
	Headline        *string
	About           *string
	Location        *string
	YearsExperience *float32
	SalaryMin       *int
	SalaryMax       *int
	SalaryCurrency   *string
	SpecializationID *string
	SoftSkills       []string
}

func (s *Service) GetMyProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	p, err := s.profiles.GetByUserID(ctx, userID)
	if errors.Is(err, domain.ErrProfileNotFound) {
		return &domain.Profile{UserID: userID, UpdatedAt: time.Now(), SoftSkills: []string{}}, nil
	}
	return p, err
}

func (s *Service) UpdateMyProfile(ctx context.Context, userID string, in UpdateProfileInput) (*domain.Profile, error) {
	if in.SalaryMin != nil && in.SalaryMax != nil && *in.SalaryMin > *in.SalaryMax {
		return nil, domain.ErrSalaryRange
	}
	if in.YearsExperience != nil && *in.YearsExperience < 0 {
		return nil, domain.ErrInvalidInput
	}

	var out *domain.Profile
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		p, err := s.profiles.GetByUserID(ctx, userID)
		if err != nil {
			// создаём профиль-заглушку при первом апдейте
			p = &domain.Profile{UserID: userID}
			applyProfileUpdate(p, in)
			if err := s.profiles.Create(ctx, p); err != nil {
				return err
			}
			out = p
			return nil
		}
		applyProfileUpdate(p, in)
		if err := s.profiles.Update(ctx, p); err != nil {
			return err
		}
		out = p
		return nil
	})
	if out != nil {
		if out.SoftSkills == nil {
			out.SoftSkills = []string{}
		}
		if out.UpdatedAt.IsZero() {
			out.UpdatedAt = time.Now()
		}
	}
	return out, err
}

func applyProfileUpdate(p *domain.Profile, in UpdateProfileInput) {
	if in.FirstName != nil {
		p.FirstName = in.FirstName
	}
	if in.LastName != nil {
		p.LastName = in.LastName
	}
	if in.MiddleName != nil {
		p.MiddleName = in.MiddleName
	}
	if in.Headline != nil {
		p.Headline = in.Headline
	}
	if in.About != nil {
		p.About = in.About
	}
	if in.Location != nil {
		p.Location = in.Location
	}
	if in.YearsExperience != nil {
		p.YearsExperience = in.YearsExperience
	}
	if in.SalaryMin != nil {
		p.SalaryMin = in.SalaryMin
	}
	if in.SalaryMax != nil {
		p.SalaryMax = in.SalaryMax
	}
	if in.SalaryCurrency != nil {
		p.SalaryCurrency = in.SalaryCurrency
	}
	if in.SpecializationID != nil {
		if *in.SpecializationID == "" {
			p.SpecializationID = nil
		} else {
			p.SpecializationID = in.SpecializationID
		}
	}
	if in.SoftSkills != nil {
		p.SoftSkills = in.SoftSkills
	}
}
