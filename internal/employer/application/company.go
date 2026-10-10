package application

import (
	"context"
	"fmt"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/domain"
)

type UpsertCompanyInput struct {
	Name            string
	Description     *string
	Industry        *string
	Website         *string
	Size            *string
	ContactPerson   *string
	ContactEmail    *string
	ContactPhone    *string
	ContactTelegram *string
}

func (s *Service) GetMyCompany(ctx context.Context, ownerID string) (*domain.Company, error) {
	return s.companies.GetByOwner(ctx, ownerID)
}

func (s *Service) UpsertMyCompany(ctx context.Context, ownerID string, in UpsertCompanyInput) (*domain.Company, error) {
	if in.Name == "" {
		return nil, domain.ErrInvalidInput
	}
	var out *domain.Company
	err := s.txm.WithinTransaction(ctx, func(ctx context.Context) error {
		c, err := s.companies.GetByOwner(ctx, ownerID)
		if err != nil && err != domain.ErrCompanyNotFound {
			return err
		}
		if err == domain.ErrCompanyNotFound {
			c = &domain.Company{OwnerUserID: ownerID}
			applyCompanyInput(c, in)
			if err := s.companies.Create(ctx, c); err != nil {
				return err
			}
			out = c
			return nil
		}
		applyCompanyInput(c, in)
		if err := s.companies.Update(ctx, c); err != nil {
			return fmt.Errorf("update company: %w", err)
		}
		out = c
		return nil
	})
	return out, err
}

func applyCompanyInput(c *domain.Company, in UpsertCompanyInput) {
	c.Name = in.Name
	c.Description = in.Description
	c.Industry = in.Industry
	c.Website = in.Website
	c.Size = in.Size
	c.ContactPerson = in.ContactPerson
	c.ContactEmail = in.ContactEmail
	c.ContactPhone = in.ContactPhone
	c.ContactTelegram = in.ContactTelegram
}
