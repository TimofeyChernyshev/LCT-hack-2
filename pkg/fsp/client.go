package fsp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client can query a remote FSP Registry service or fall back to local mock registry
type Client struct {
	baseURL    string
	httpClient *http.Client
	fallback   *MockRegistry
}

// NewClient creates a registry client with optional remote URL and in-memory mock fallback
func NewClient(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
		fallback: NewMockRegistry(),
	}
}

// GetMember retrieves member details and achievements
func (c *Client) GetMember(ctx context.Context, fspID string) (*Member, error) {
	if c.baseURL != "" {
		url := fmt.Sprintf("%s/api/v1/fsp/registry/members/%s", c.baseURL, fspID)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err == nil {
			resp, err := c.httpClient.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var member Member
					if err := json.NewDecoder(resp.Body).Decode(&member); err == nil {
						return &member, nil
					}
				} else if resp.StatusCode == http.StatusNotFound {
					return nil, ErrMemberNotFound
				}
			}
		}
	}

	// Fallback to local mock registry
	return c.fallback.GetMember(ctx, fspID)
}

// SearchMembers searches the registry
func (c *Client) SearchMembers(ctx context.Context, query string, rank SportsRank, region string, limit, offset int) ([]Member, int, error) {
	if c.baseURL != "" {
		url := fmt.Sprintf("%s/api/v1/fsp/registry/members?q=%s&rank=%s&region=%s&limit=%d&offset=%d",
			c.baseURL, query, rank, region, limit, offset)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err == nil {
			resp, err := c.httpClient.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var result struct {
						Members []Member `json:"members"`
						Total   int      `json:"total"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
						return result.Members, result.Total, nil
					}
				}
			}
		}
	}

	return c.fallback.SearchMembers(ctx, query, rank, region, limit, offset)
}

// VerifyMember validates credentials against the registry
func (c *Client) VerifyMember(ctx context.Context, req VerificationRequest) (*VerificationResult, error) {
	if c.baseURL != "" {
		url := fmt.Sprintf("%s/api/v1/fsp/registry/verify", c.baseURL)
		body, _ := json.Marshal(req)
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err == nil {
			httpReq.Header.Set("Content-Type", "application/json")
			resp, err := c.httpClient.Do(httpReq)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var res VerificationResult
					if err := json.NewDecoder(resp.Body).Decode(&res); err == nil {
						return &res, nil
					}
				}
			}
		}
	}

	return c.fallback.VerifyMember(ctx, req)
}

// FallbackRegistry returns the internal mock registry instance (useful for adding custom test data)
func (c *Client) FallbackRegistry() *MockRegistry {
	return c.fallback
}
