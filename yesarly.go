package main

// https://docs.esa.io/posts/102

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// EsaClient esa.io api client
type EsaClient struct {
	token    string
	team     string
	endpoint *url.URL
	client   *http.Client
}

// Post esa post
type Post struct {
	URL       string    `json:"url"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAT time.Time `json:"updated_at"`
	CreatedBy struct {
		ScreenName string `json:"screen_name"`
		Name       string `json:"name"`
	} `json:"created_by"`
}

// RetryInfo retry info
type RetryInfo struct {
	RateLimitExceeded bool
	RetryAfter        int
}

// PostsResponse /v1/teams/:team_name/posts response
type PostsResponse struct {
	MaxPerPage int `json:"max_per_page"`
	PerPage    int `json:"per_page"`
	Page       int `json:"page"`
	TotalCount int `json:"total_count"`
	NextPage   int `json:"next_page"`
	PrevPage   int `json:"prev_page"`
	RetryInfo  RetryInfo
	Posts      []Post `json:"posts"`
}

func createSearchQueryString(username string, from, to time.Time, page int) string {
	// + sign must not be encoded, other signs must be properly encoded
	// https://play.golang.org/p/pOfrn-Wsq5
	//   * `>` = `%3E`
	//   * `<` = `%3C`
	//   * `+` = `%2B`
	//   * `:` = `%3A`
	pUsername := fmt.Sprintf("user%%3A%s", username)
	pFrom := fmt.Sprintf("created%%3A%%3E%s", from.Format("2006-01-02"))
	pTo := fmt.Sprintf("created%%3A%%3C%s", to.Format("2006-01-02"))
	qs := fmt.Sprintf("?q=%s+%s+%s", pUsername, pFrom, pTo)
	if page > 1 {
		pqs := fmt.Sprintf("&page=%d", page)
		qs = qs + pqs
	}
	return qs
}

// GetPostsByUsername get posts by username
// https://docs.esa.io/posts/104
func (c *EsaClient) GetPostsByUsername(username string, from, to time.Time, page int) (*PostsResponse, error) {
	p := c.endpoint.String() + fmt.Sprintf("/v1/teams/%s/posts", c.team)
	qs := createSearchQueryString(username, from, to, page)
	req, err := http.NewRequest(http.MethodGet, p+qs, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusTooManyRequests {
		return nil, errors.New(fmt.Sprintf("esa status code %d", resp.StatusCode))
	}

	response := &PostsResponse{}

	// retry
	retryInfo := resp.Header.Values("Retry-After")
	if len(retryInfo) != 0 {
		ra, err := strconv.Atoi(retryInfo[0])
		if err != nil {
			return nil, err
		}
		ri := RetryInfo{
			RateLimitExceeded: true,
			RetryAfter:        ra,
		}
		response.RetryInfo = ri
		return response, nil
	}

	// no retry
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(response); err != nil {
		return nil, err
	}
	return response, nil
}

// GetPostsByUsernameWithRetry GetPostsByUsername with retry
func (c *EsaClient) GetPostsByUsernameWithRetry(username string, from, to time.Time, page int) (*PostsResponse, error) {
	resp, err := c.GetPostsByUsername(username, from, to, page)
	if err != nil {
		return nil, err
	}
	if !resp.RetryInfo.RateLimitExceeded {
		return resp, err
	}

	log.Printf("retry after %d sec", resp.RetryInfo.RetryAfter)
	time.Sleep(time.Duration(resp.RetryInfo.RetryAfter))
	resp2, err := c.GetPostsByUsername(username, from, to, page)
	if err != nil {
		return nil, err
	}
	return resp2, nil
}

// NewEsaClient creates new esa client
func NewEsaClient(team, endpoint, token string, timeoutSec time.Duration) (*EsaClient, error) {
	c := &http.Client{
		Timeout: time.Second * timeoutSec,
	}
	e, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return &EsaClient{
		team:     team,
		token:    token,
		endpoint: e,
		client:   c,
	}, nil
}
