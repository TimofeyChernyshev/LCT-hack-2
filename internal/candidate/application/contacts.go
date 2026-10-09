package application

import (
	"context"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
)

func (s *Service) GetMyContacts(ctx context.Context, userID string) (*domain.Contacts, error) {
	c, err := s.contacts.GetByUserID(ctx, userID)
	if err != nil {
		return &domain.Contacts{}, nil // пустые — норм для нового пользователя
	}
	return c, nil
}

func (s *Service) ReplaceMyContacts(ctx context.Context, userID string, c *domain.Contacts) error {
	return s.contacts.Upsert(ctx, userID, c)
}
