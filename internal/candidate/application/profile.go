package application

import (
	"context"

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
	SalaryCurrency  *string
	SoftSkills      []string
}

func (s *Service) GetMyProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	return s.profiles.GetByUserID(ctx, userID)
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
	if in.SoftSkills != nil {
		p.SoftSkills = in.SoftSkills
	}
}
