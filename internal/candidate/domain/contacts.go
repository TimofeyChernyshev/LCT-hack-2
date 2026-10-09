package domain

type Contacts struct {
	Email    string
	Phone    *string
	Telegram *string
	GitHub   *string
	LinkedIn *string
	Website  *string
}

// MaskedContacts — результат маскирования. Флаги показывают, что скрыто.
type MaskedContacts struct {
	Contacts
	Masked       bool
	MaskedFields []string
}

// Mask возвращает копию контактов, где каждое поле заменено маской.
// Если поле пустое — оставляем пустым и не включаем в maskedFields.
func (c Contacts) Mask() MaskedContacts {
	out := MaskedContacts{
		Contacts: Contacts{Email: maskEmail(c.Email)},
		Masked:   true,
	}
	if c.Email != "" {
		out.MaskedFields = append(out.MaskedFields, "email")
	}
	out.Phone = maskPtr(c.Phone, "phone", &out.MaskedFields, maskPhone)
	out.Telegram = maskPtr(c.Telegram, "telegram", &out.MaskedFields, maskGeneric)
	out.GitHub = maskPtr(c.GitHub, "github", &out.MaskedFields, maskGeneric)
	out.LinkedIn = maskPtr(c.LinkedIn, "linkedin", &out.MaskedFields, maskGeneric)
	out.Website = maskPtr(c.Website, "website", &out.MaskedFields, maskGeneric)
	return out
}

func maskPtr(v *string, name string, fields *[]string, fn func(string) string) *string {
	if v == nil || *v == "" {
		return nil
	}
	m := fn(*v)
	*fields = append(*fields, name)
	return &m
}

func maskEmail(s string) string {
	at := -1
	for i, r := range s {
		if r == '@' {
			at = i
			break
		}
	}
	if at <= 0 {
		return "***"
	}
	local := s[:at]
	domain := s[at+1:]
	return string(local[0]) + "***@" + domain
}

func maskPhone(s string) string {
	// оставляем последние 2 цифры
	digits := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}
	if len(digits) < 2 {
		return "***"
	}
	return "***" + string(digits[len(digits)-2:])
}

func maskGeneric(string) string { return "***" }
