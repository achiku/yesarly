package main

// https://docs.esa.io/posts/102

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

// PostsResponse /v1/teams/:team_name/posts response
type PostsResponse struct {
	MaxPerPage int    `json:"max_per_page"`
	PerPage    int    `json:"per_page"`
	Page       int    `json:"page"`
	TotalCount int    `json:"total_count"`
	NextPage   int    `json:"next_page"`
	PrevPage   int    `json:"prev_page"`
	Posts      []Post `json:"posts"`
}

// GetPostsByUsername get posts by username
// https://docs.esa.io/posts/104
func (c *EsaClient) GetPostsByUsername(username string, from, to time.Time, page int) (*PostsResponse, error) {
	p := c.endpoint.String() + fmt.Sprintf("/v1/teams/%s/posts", c.team)
	// + sign must not be encoded, other signs must be properly encoded
	//   * `>` = `%3E`
	//   * `<` = `%3C`
	//   * `+` = `%2B`
	//   * `:` = `%3A`
	un := fmt.Sprintf("user%%3A%s", username)
	ft := fmt.Sprintf("created%%3A%%3E%s", "2020-10-01")
	tt := fmt.Sprintf("created%%3A%%3C%s", "2020-12-15")
	qs := fmt.Sprintf("?q=%s+%s+%s", un, ft, tt)
	if page > 1 {
		pqs := fmt.Sprintf("&page=%d", page)
		qs = qs + pqs
	}
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
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("esa status code %d", resp.StatusCode))
	}
	decoder := json.NewDecoder(resp.Body)
	response := &PostsResponse{}
	if err := decoder.Decode(response); err != nil {
		return nil, err
	}
	return response, nil
}

// NewEsaClient creates new esa client
func NewEsaClient(team, token string, timeoutSec time.Duration) (*EsaClient, error) {
	c := &http.Client{
		Timeout: time.Second * timeoutSec,
	}
	u, err := url.Parse("https://api.esa.io/")
	if err != nil {
		return nil, err
	}
	return &EsaClient{
		team:     team,
		token:    token,
		endpoint: u,
		client:   c,
	}, nil
}
