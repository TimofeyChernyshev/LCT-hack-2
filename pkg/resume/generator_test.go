package resume

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCandidatePDF(t *testing.T) {
	profile := CandidateExportProfile{
		UserID:             uuid.New(),
		FullName:           "Александр Смирнов",
		Headline:           "Senior Backend Developer (Go)",
		SpecializationName: "Backend",
		GradeName:          "Senior",
		GradeRank:          3,
		Location:           "Москва",
		TestScore:          96.5,
		TestPercentile:     98.0,
		HasFSP:             true,
		SportsRank:         "Мастер спорта",
		FSPRating:          2480,
		FSPAchievements:    []string{"1 место Чемпионат ФСП 2024", "серебряный призёр Кубка ФСП"},
		Skills:             []string{"Go", "PostgreSQL", "Redis", "Docker", "Kubernetes"},
		YearsExperience:    6.5,
		MaskContacts:       true, // Masked privacy state
		Email:              "alex@example.com",
		Phone:              "+7 999 123-45-67",
		Telegram:           "@alex_smirnov",
		GeneratedAt:        time.Now(),
	}

	pdfBytes, err := GenerateCandidatePDF(profile)
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)

	// Check PDF structure
	assert.True(t, strings.HasPrefix(string(pdfBytes), "%PDF-1.4"))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(string(pdfBytes)), "%%EOF"))

	// Check content presence in PDF objects
	pdfContent := string(pdfBytes)
	assert.Contains(t, pdfContent, "FSP PROFILE")
	assert.Contains(t, pdfContent, "Aleksandr Smirnov")
	assert.Contains(t, pdfContent, "SPORTS PROGRAMMING TRACK")
	assert.Contains(t, pdfContent, "CONTACTS MASKED")

	// Round-trip text extraction from PDF
	extractedText, err := ExtractTextFromPDF(pdfBytes)
	require.NoError(t, err)
	require.NotEmpty(t, extractedText)
	assert.Contains(t, extractedText, "FSP PROFILE")
}

func TestGenerateCandidatePDFUnlockedContacts(t *testing.T) {
	profile := CandidateExportProfile{
		UserID:             uuid.New(),
		FullName:           "Дмитрий Ковалев",
		SpecializationName: "Backend Python",
		GradeName:          "Middle",
		TestScore:          85.0,
		HasFSP:             false,
		Skills:             []string{"Python", "FastAPI", "PostgreSQL"},
		YearsExperience:    3.0,
		MaskContacts:       false, // Unlocked state
		Email:              "dmitry@example.com",
		Telegram:           "@dmitry_dev",
		GeneratedAt:        time.Now(),
	}

	pdfBytes, err := GenerateCandidatePDF(profile)
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)

	pdfContent := string(pdfBytes)
	assert.Contains(t, pdfContent, "CONTACTS UNLOCKED")
	assert.Contains(t, pdfContent, "dmitry@example.com")
}
