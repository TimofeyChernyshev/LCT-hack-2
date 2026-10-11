package application

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
)

type dummyProfileRepo struct {
	profile domain.Profile
}

func (d *dummyProfileRepo) GetByUserID(_ context.Context, _ string) (*domain.Profile, error) {
	return &d.profile, nil
}
func (d *dummyProfileRepo) Create(_ context.Context, _ *domain.Profile) error { return nil }
func (d *dummyProfileRepo) Update(_ context.Context, _ *domain.Profile) error { return nil }

type dummyVisibilityRepo struct {
	vis domain.Visibility
}

func (d *dummyVisibilityRepo) Get(_ context.Context, _ string) (*domain.Visibility, error) {
	return &d.vis, nil
}
func (d *dummyVisibilityRepo) Upsert(_ context.Context, _ string, _ *domain.Visibility) error {
	return nil
}

type dummyContactsRepo struct {
	contacts domain.Contacts
}

func (d *dummyContactsRepo) GetByUserID(_ context.Context, _ string) (*domain.Contacts, error) {
	return &d.contacts, nil
}
func (d *dummyContactsRepo) Upsert(_ context.Context, _ string, _ *domain.Contacts) error {
	return nil
}

type dummyRevealRepo struct {
	revealed bool
}

func (d *dummyRevealRepo) IsRevealed(_ context.Context, _, _ string) (bool, error) {
	return d.revealed, nil
}
func (d *dummyRevealRepo) Create(_ context.Context, _ *domain.ContactReveal) error { return nil }

type dummyFSPRepo struct{}

func (d *dummyFSPRepo) GetMemberID(_ context.Context, _ string) (*string, *time.Time, error) {
	return nil, nil, nil
}
func (d *dummyFSPRepo) LinkMemberID(_ context.Context, _, _ string) error { return nil }
func (d *dummyFSPRepo) ListAchievements(_ context.Context, _ string) ([]domain.FSPAchievement, error) {
	return nil, nil
}
func (d *dummyFSPRepo) CountAchievements(_ context.Context, _ string) (int, *int, error) {
	return 2, nil, nil
}
func (d *dummyFSPRepo) CreateAchievement(_ context.Context, _ *domain.FSPAchievement) error {
	return nil
}

func strPtr(s string) *string {
	return &s
}

func TestBuildCardView_MaskedWhenNotRevealed(t *testing.T) {
	phone := "+7 (999) 111-22-33"
	tg := "@candidate"
	salaryMin := 200000
	salaryMax := 300000

	svc := &Service{
		profiles: &dummyProfileRepo{
			profile: domain.Profile{
				UserID:    "cand-1",
				FirstName: strPtr("Иван"),
				LastName:  strPtr("Иванов"),
				SalaryMin: &salaryMin,
				SalaryMax: &salaryMax,
			},
		},
		visibility: &dummyVisibilityRepo{
			vis: domain.Visibility{
				Contacts:   true,
				Links:      true,
				FSP:        true,
				Experience: true,
				Salary:     true,
				SoftSkills: true,
			},
		},
		contacts: &dummyContactsRepo{
			contacts: domain.Contacts{
				Email:    "ivan@company.ru",
				Phone:    &phone,
				Telegram: &tg,
			},
		},
		reveals: &dummyRevealRepo{revealed: false},
		fsp:     &dummyFSPRepo{},
	}

	card, err := svc.BuildCardView(context.Background(), "cand-1", "emp-1")
	require.NoError(t, err)
	require.NotNil(t, card)

	// DoD verification: Contacts are masked
	assert.True(t, card.Contacts.Masked, "contacts must be masked before reveal")
	assert.Equal(t, "i***@company.ru", card.Contacts.Email)
	assert.Equal(t, "***33", *card.Contacts.Phone)
	assert.Equal(t, "***", *card.Contacts.Telegram)
	assert.False(t, card.Salary.Masked)
}

func TestBuildCardView_UnmaskedWhenRevealed(t *testing.T) {
	phone := "+7 (999) 111-22-33"
	tg := "@candidate"

	svc := &Service{
		profiles: &dummyProfileRepo{
			profile: domain.Profile{UserID: "cand-1", FirstName: strPtr("Иван")},
		},
		visibility: &dummyVisibilityRepo{
			vis: domain.DefaultVisibility(),
		},
		contacts: &dummyContactsRepo{
			contacts: domain.Contacts{
				Email:    "ivan@company.ru",
				Phone:    &phone,
				Telegram: &tg,
			},
		},
		reveals: &dummyRevealRepo{revealed: true},
		fsp:     &dummyFSPRepo{},
	}

	card, err := svc.BuildCardView(context.Background(), "cand-1", "emp-1")
	require.NoError(t, err)
	require.NotNil(t, card)

	// DoD verification: Contacts are unmasked after accepted reveal
	assert.False(t, card.Contacts.Masked, "contacts must be revealed after accepted invitation")
	assert.Equal(t, "ivan@company.ru", card.Contacts.Email)
	assert.Equal(t, "+7 (999) 111-22-33", *card.Contacts.Phone)
	assert.Equal(t, "@candidate", *card.Contacts.Telegram)
}

func TestBuildCardView_SalaryMaskedWhenDisabled(t *testing.T) {
	salaryMin := 200000
	salaryMax := 300000

	vis := domain.DefaultVisibility()
	vis.Salary = false

	svc := &Service{
		profiles: &dummyProfileRepo{
			profile: domain.Profile{
				UserID:    "cand-1",
				SalaryMin: &salaryMin,
				SalaryMax: &salaryMax,
			},
		},
		visibility: &dummyVisibilityRepo{vis: vis},
		contacts:   &dummyContactsRepo{},
		reveals:    &dummyRevealRepo{revealed: true},
		fsp:        &dummyFSPRepo{},
	}

	card, err := svc.BuildCardView(context.Background(), "cand-1", "emp-1")
	require.NoError(t, err)
	assert.True(t, card.Salary.Masked)
	assert.Nil(t, card.Salary.Min)
	assert.Nil(t, card.Salary.Max)
}
