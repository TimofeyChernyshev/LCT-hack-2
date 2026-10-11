package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContacts_Mask(t *testing.T) {
	phone := "+7 (999) 123-45-67"
	tg := "@candidate_tg"
	gh := "https://github.com/candidate"
	li := "https://linkedin.com/in/candidate"
	web := "https://candidate.dev"

	c := Contacts{
		Email:    "john.doe@techcorp.ru",
		Phone:    &phone,
		Telegram: &tg,
		GitHub:   &gh,
		LinkedIn: &li,
		Website:  &web,
	}

	masked := c.Mask()

	assert.True(t, masked.Masked)
	assert.Contains(t, masked.MaskedFields, "email")
	assert.Contains(t, masked.MaskedFields, "phone")
	assert.Contains(t, masked.MaskedFields, "telegram")
	assert.Contains(t, masked.MaskedFields, "github")
	assert.Contains(t, masked.MaskedFields, "linkedin")
	assert.Contains(t, masked.MaskedFields, "website")

	assert.Equal(t, "j***@techcorp.ru", masked.Email)
	assert.NotNil(t, masked.Phone)
	assert.Equal(t, "***67", *masked.Phone)
	assert.NotNil(t, masked.Telegram)
	assert.Equal(t, "***", *masked.Telegram)
	assert.NotNil(t, masked.GitHub)
	assert.Equal(t, "***", *masked.GitHub)
	assert.NotNil(t, masked.LinkedIn)
	assert.Equal(t, "***", *masked.LinkedIn)
	assert.NotNil(t, masked.Website)
	assert.Equal(t, "***", *masked.Website)
}

func TestContacts_Mask_EmptyFields(t *testing.T) {
	c := Contacts{
		Email: "",
	}

	masked := c.Mask()

	assert.True(t, masked.Masked)
	assert.Empty(t, masked.MaskedFields)
	assert.Equal(t, "***", masked.Email)
	assert.Nil(t, masked.Phone)
	assert.Nil(t, masked.Telegram)
	assert.Nil(t, masked.GitHub)
	assert.Nil(t, masked.LinkedIn)
	assert.Nil(t, masked.Website)
}

func TestContacts_Mask_PhoneShort(t *testing.T) {
	phone := "1"
	c := Contacts{
		Email: "a@b.com",
		Phone: &phone,
	}

	masked := c.Mask()
	assert.NotNil(t, masked.Phone)
	assert.Equal(t, "***", *masked.Phone)
}
