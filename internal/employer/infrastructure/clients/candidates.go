package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/application"
)

type CandidateClient struct {
	baseURL string
	http    *http.Client
}

func NewCandidateClient(baseURL string, timeout time.Duration) *CandidateClient {
	return &CandidateClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

type candidateCardDTO struct {
	UserID           string   `json:"userId"`
	DisplayName      string   `json:"displayName"`
	Headline         *string  `json:"headline"`
	CategoryID       *string  `json:"categoryId"`
	GradeID          *string  `json:"gradeId"`
	SpecializationID *string  `json:"specializationId"`
	YearsExperience  *float32 `json:"yearsExperience"`
	Location         *string  `json:"location"`
	SoftSkills       []string `json:"softSkills"`
	Salary           *struct {
		Min      *int    `json:"min"`
		Max      *int    `json:"max"`
		Currency *string `json:"currency"`
		Masked   bool    `json:"masked"`
	} `json:"salary"`
	Fsp *struct {
		HasFSP            bool `json:"hasFSP"`
		AchievementsCount int  `json:"achievementsCount"`
		BestPlace         *int `json:"bestPlace"`
	} `json:"fsp"`
	Contacts *struct {
		Email        *string  `json:"email"`
		Phone        *string  `json:"phone"`
		Telegram     *string  `json:"telegram"`
		GitHub       *string  `json:"github"`
		LinkedIn     *string  `json:"linkedin"`
		Website      *string  `json:"website"`
		Masked       bool     `json:"masked"`
		MaskedFields []string `json:"maskedFields"`
	} `json:"contacts"`
}

func (c *CandidateClient) GetCard(ctx context.Context, authHeader, candidateUserID string) (*application.CandidateCard, error) {
	url := c.baseURL + "/candidates/" + candidateUserID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("candidate request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return nil, application.ErrNotFound
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("candidate: %d %s", resp.StatusCode, string(body))
	}

	var dto candidateCardDTO
	if err := json.Unmarshal(body, &dto); err != nil {
		return nil, fmt.Errorf("decode candidate card: %w", err)
	}

	card := &application.CandidateCard{
		UserID:           dto.UserID,
		DisplayName:      dto.DisplayName,
		Headline:         dto.Headline,
		CategoryID:       dto.CategoryID,
		GradeID:          dto.GradeID,
		SpecializationID: dto.SpecializationID,
		YearsExperience:  dto.YearsExperience,
		Location:         dto.Location,
		SoftSkills:       dto.SoftSkills,
	}
	if dto.Salary != nil {
		card.Salary = &application.CardSalary{
			Min:      dto.Salary.Min,
			Max:      dto.Salary.Max,
			Currency: dto.Salary.Currency,
			Masked:   dto.Salary.Masked,
		}
	}
	if dto.Fsp != nil {
		card.FSP = &application.CardFSP{
			HasFSP:            dto.Fsp.HasFSP,
			AchievementsCount: dto.Fsp.AchievementsCount,
			BestPlace:         dto.Fsp.BestPlace,
		}
	}
	if dto.Contacts != nil {
		card.Contacts = &application.CardContacts{
			Email:        dto.Contacts.Email,
			Phone:        dto.Contacts.Phone,
			Telegram:     dto.Contacts.Telegram,
			GitHub:       dto.Contacts.GitHub,
			LinkedIn:     dto.Contacts.LinkedIn,
			Website:      dto.Contacts.Website,
			Masked:       dto.Contacts.Masked,
			MaskedFields: dto.Contacts.MaskedFields,
		}
	}
	return card, nil
}
