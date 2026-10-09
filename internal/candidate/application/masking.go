package application

import (
	"context"
	"fmt"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
)

// CardView — то, что возвращается работодателю. Всегда с маской контактов,
// если reveal не подтверждён.
type CardView struct {
	Profile    domain.Profile
	Contacts   domain.MaskedContacts
	Visibility domain.Visibility
	FSP        CardFSP
	Salary     CardSalary
}

type CardFSP struct {
	HasFSP            bool
	AchievementsCount int
	BestPlace         *int
}

type CardSalary struct {
	Min      *int
	Max      *int
	Currency string
	Masked   bool
}

// BuildCardView - единственное место, где решается, что видно работодателю.
// Правило: контакты раскрываются только если в БД есть запись contact_reveals
// между этой парой. Никаких query-параметров, никаких заголовков.
func (s *Service) BuildCardView(ctx context.Context, candidateUserID, employerUserID string) (*CardView, error) {
	p, err := s.profiles.GetByUserID(ctx, candidateUserID)
	if err != nil {
		return nil, err
	}
	vis, err := s.visibility.Get(ctx, candidateUserID)
	if err != nil {
		// fallback на дефолт, если настройки ещё не задавали
		v := domain.DefaultVisibility()
		vis = &v
	}
	c, err := s.contacts.GetByUserID(ctx, candidateUserID)
	if err != nil {
		c = &domain.Contacts{}
	}

	revealed, err := s.reveals.IsRevealed(ctx, candidateUserID, employerUserID)
	if err != nil {
		return nil, fmt.Errorf("check reveal: %w", err)
	}

	// показывать = revealed И пользователь разрешил contacts
	showContacts := revealed && vis.Contacts

	var masked domain.MaskedContacts
	if showContacts {
		masked = domain.MaskedContacts{Contacts: *c}
	} else {
		masked = c.Mask()
	}

	// Ссылки (github/linkedin/website) - отдельная политика, если vis.Links=false
	if !vis.Links {
		masked.GitHub = nil
		masked.LinkedIn = nil
		masked.Website = nil
	}

	view := &CardView{
		Profile:    *p,
		Contacts:   masked,
		Visibility: *vis,
	}

	// FSP
	if vis.FSP {
		count, best, err := s.fsp.CountAchievements(ctx, candidateUserID)
		if err == nil && count > 0 {
			view.FSP = CardFSP{HasFSP: true, AchievementsCount: count, BestPlace: best}
		}
	}

	// Salary - тоже может скрываться
	view.Salary = CardSalary{
		Min:      p.SalaryMin,
		Max:      p.SalaryMax,
		Currency: strOr(p.SalaryCurrency, "RUB"),
	}
	if !vis.Salary {
		view.Salary.Min = nil
		view.Salary.Max = nil
		view.Salary.Masked = true
	}

	return view, nil
}

func strOr(s *string, def string) string {
	if s == nil || *s == "" {
		return def
	}
	return *s
}
