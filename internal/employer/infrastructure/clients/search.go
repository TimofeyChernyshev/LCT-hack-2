package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/application"
)

type SearchClient struct {
	baseURL string
	http    *http.Client
}

func NewSearchClient(baseURL string, timeout time.Duration) *SearchClient {
	return &SearchClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

type searchPageDTO struct {
	Items  []searchHitDTO `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type searchHitDTO struct {
	UserID               string   `json:"userId"`
	DisplayName          string   `json:"displayName"`
	CategoryID           string   `json:"categoryId"`
	GradeID              string   `json:"gradeId"`
	SpecializationID     string   `json:"specializationId"`
	TestScore            *float32 `json:"testScore"`
	FSPAchievementsCount int      `json:"fspAchievementsCount"`
	FSPBestPlace         *int     `json:"fspBestPlace"`
	FSPWeightSum         int      `json:"fspWeightSum"`
	YearsExperience      *float32 `json:"yearsExperience"`
	Score                float32  `json:"score"`
	Reasons              []string `json:"reasons"`
}

func (c *SearchClient) SearchCandidates(ctx context.Context, authHeader string, q application.SearchQuery) (*application.SearchPage, error) {
	u, _ := url.Parse(c.baseURL + "/candidates")
	vals := u.Query()
	if q.CategoryID != nil {
		vals.Set("categoryId", *q.CategoryID)
	}
	if q.SpecializationID != nil {
		vals.Set("specializationId", *q.SpecializationID)
	}
	if q.GradeID != nil {
		vals.Set("gradeId", *q.GradeID)
	}
	for _, s := range q.Stack {
		vals.Add("stack", s)
	}
	if q.HasFSP != nil {
		vals.Set("hasFsp", strconv.FormatBool(*q.HasFSP))
	}
	if q.MinYearsExperience != nil {
		vals.Set("minYearsExperience", strconv.FormatFloat(float64(*q.MinYearsExperience), 'f', -1, 64))
	}
	if q.Location != nil {
		vals.Set("location", *q.Location)
	}
	if q.Sort != "" {
		vals.Set("sort", q.Sort)
	}
	vals.Set("limit", strconv.Itoa(q.Limit))
	vals.Set("offset", strconv.Itoa(q.Offset))
	u.RawQuery = vals.Encode()

	return c.doSearch(ctx, authHeader, u.String())
}

func (c *SearchClient) MatchesForNeed(ctx context.Context, authHeader, needID string, q application.MatchesQuery) (*application.SearchPage, error) {
	u, _ := url.Parse(c.baseURL + "/needs/" + needID + "/matches")
	vals := u.Query()
	vals.Set("limit", strconv.Itoa(q.Limit))
	vals.Set("offset", strconv.Itoa(q.Offset))
	u.RawQuery = vals.Encode()
	return c.doSearch(ctx, authHeader, u.String())
}

func (c *SearchClient) doSearch(ctx context.Context, authHeader, fullURL string) (*application.SearchPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("search: %d %s", resp.StatusCode, string(body))
	}

	var dto searchPageDTO
	if err := json.Unmarshal(body, &dto); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	page := &application.SearchPage{
		Total:  dto.Total,
		Limit:  dto.Limit,
		Offset: dto.Offset,
	}
	for _, h := range dto.Items {
		page.Items = append(page.Items, application.SearchHit{
			UserID:               h.UserID,
			DisplayName:          h.DisplayName,
			CategoryID:           h.CategoryID,
			GradeID:              h.GradeID,
			SpecializationID:     h.SpecializationID,
			TestScore:            h.TestScore,
			FSPAchievementsCount: h.FSPAchievementsCount,
			FSPBestPlace:         h.FSPBestPlace,
			FSPWeightSum:         h.FSPWeightSum,
			YearsExperience:      h.YearsExperience,
			Score:                h.Score,
			Reasons:              h.Reasons,
		})
	}
	return page, nil
}
