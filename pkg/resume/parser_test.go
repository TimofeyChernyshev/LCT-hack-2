package resume

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseResumeFull(t *testing.T) {
	resumeText := `
Александр Дмитриевич Смирнов
Senior Backend Developer (Go / Python)
г. Москва

Контакты:
Email: alex.smirnov@example.com
Телефон: +7 (999) 123-45-67
Telegram: @alex_backend
GitHub: github.com/alex-smirnov
LinkedIn: linkedin.com/in/alex-smirnov-dev

Опыт работы: 6 лет
2021 — настоящее время
Яндекс Финтех
Senior Go Разработчик
Разработка высоконагруженных микросервисов платежной системы.
Стек: Go, PostgreSQL, Redis, Docker, Kubernetes, gRPC, Kafka.

2018 — 2021
VK Tech
Backend-разработчик
Стек: Python, Django, PostgreSQL, Docker, Redis.

Образование:
МГТУ им. Баумана, 2018, Информатика и вычислительная техника

Достижения в спортивном программировании:
Участник Федерации спортивного программирования (ФСП).
ID: FSP-RU-77-00101
Спортивное звание: Мастер спорта
Серебряный призёр Кубка ФСП 2025
`

	parsed := ParseResume(resumeText)

	// Check name
	assert.Equal(t, "Александр Дмитриевич Смирнов", parsed.FullName)
	assert.Equal(t, "Александр", parsed.FirstName)
	assert.Equal(t, "Смирнов", parsed.LastName)
	assert.Equal(t, "Дмитриевич", parsed.MiddleName)

	// Check contacts
	assert.Equal(t, "alex.smirnov@example.com", parsed.Contacts.Email)
	assert.Equal(t, "+7 (999) 123-45-67", parsed.Contacts.Phone)
	assert.Equal(t, "@alex_backend", parsed.Contacts.Telegram)
	assert.Equal(t, "https://github.com/alex-smirnov", parsed.Contacts.GitHub)
	assert.Equal(t, "https://linkedin.com/in/alex-smirnov-dev", parsed.Contacts.LinkedIn)

	// Check location
	assert.Equal(t, "Москва", parsed.Location)

	// Check skills
	assert.Contains(t, parsed.Skills, "Go")
	assert.Contains(t, parsed.Skills, "Python")
	assert.Contains(t, parsed.Skills, "PostgreSQL")
	assert.Contains(t, parsed.Skills, "Redis")
	assert.Contains(t, parsed.Skills, "Docker")
	assert.Contains(t, parsed.Skills, "Kubernetes")
	assert.Contains(t, parsed.Skills, "gRPC")
	assert.Contains(t, parsed.Skills, "Kafka")

	// Check experience
	assert.InDelta(t, 6.0, parsed.TotalYearsExperience, 0.5)
	require.NotEmpty(t, parsed.Experiences)

	// Check education
	require.NotEmpty(t, parsed.Education)
	assert.Equal(t, "МГТУ им. Баумана", parsed.Education[0].Institution)

	// Check FSP mentions
	require.NotNil(t, parsed.FSPMention)
	assert.True(t, parsed.FSPMention.HasFSP)
	assert.Equal(t, "FSP-RU-77-00101", parsed.FSPMention.FSPID)
	assert.Equal(t, "Мастер спорта", parsed.FSPMention.SportsRank)
}

func TestParseResumeWithoutFSP(t *testing.T) {
	resumeText := `
Елена Николаевна Васильева
Frontend React Engineer
г. Санкт-Петербург

Email: elena.react@mail.ru
Телефон: 8 (911) 555-44-33
Telegram: @elena_fe

Опыт работы: 3 года
Стек технологий: React, TypeScript, Vue, JavaScript, HTML, CSS, Docker, Webpack.
`

	parsed := ParseResume(resumeText)

	assert.Equal(t, "Елена Николаевна Васильева", parsed.FullName)
	assert.Equal(t, "elena.react@mail.ru", parsed.Contacts.Email)
	assert.Equal(t, "8 (911) 555-44-33", parsed.Contacts.Phone)
	assert.Equal(t, "@elena_fe", parsed.Contacts.Telegram)
	assert.Equal(t, "Санкт-Петербург", parsed.Location)

	assert.Contains(t, parsed.Skills, "React")
	assert.Contains(t, parsed.Skills, "TypeScript")
	assert.Contains(t, parsed.Skills, "Docker")

	assert.InDelta(t, 3.0, parsed.TotalYearsExperience, 0.5)
	assert.Nil(t, parsed.FSPMention)
}
