package domain

import "time"

type Company struct {
	ID              string
	OwnerUserID     string
	Name            string
	Description     *string
	Industry        *string
	Website         *string
	Size            *string
	ContactPerson   *string
	ContactEmail    *string
	ContactPhone    *string
	ContactTelegram *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
