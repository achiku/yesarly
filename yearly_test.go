package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"
	"time"
)

func TestGetPostsByUsername(t *testing.T) {
	u := "https://api.esa.io/"
	apikey := os.Getenv("ESA_API_KEY")
	c, err := NewEsaClient("kanmu", u, apikey, time.Duration(5))
	if err != nil {
		t.Fatal(err)
	}
	res, err := c.GetPostsByUsername("achiku", time.Now(), time.Now(), 1)
	if err != nil {
		t.Fatal(err)
	}
	posts := res.Posts
	nextPage := res.NextPage
	for nextPage != 0 {
		res, err := c.GetPostsByUsername("achiku", time.Now(), time.Now(), nextPage)
		if err != nil {
			t.Fatal(err)
		}
		posts = append(posts, res.Posts...)
		nextPage = res.NextPage

	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	for _, r := range posts {
		fmt.Printf("%s %s %s\n", r.CreatedBy.ScreenName, r.FullName, r.CreatedAt)
	}
}

func TestPostsSort(t *testing.T) {
	posts := []Post{
		Post{
			URL:       "aaaa",
			FullName:  "name",
			CreatedAt: time.Date(2020, 10, 1, 0, 0, 0, 0, time.Local),
		},
		Post{
			URL:       "aaaa",
			FullName:  "name",
			CreatedAt: time.Date(2020, 9, 8, 0, 0, 0, 0, time.Local),
		},
		Post{
			URL:       "aaaa",
			FullName:  "name",
			CreatedAt: time.Date(2020, 9, 1, 0, 0, 0, 0, time.Local),
		},
		Post{
			URL:       "aaaa",
			FullName:  "name",
			CreatedAt: time.Date(2020, 12, 1, 0, 0, 0, 0, time.Local),
		},
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})
	for _, p := range posts {
		log.Printf("%v", p)
	}
}

func NewTestEsaServer(t *testing.T, h map[string]http.Handler) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for url, handler := range h {
		mux.Handle(url, handler)
	}
	return httptest.NewServer(mux)
}

func NewTestEsaClient(t *testing.T, endpoint string) *EsaClient {
	t.Helper()
	c, err := NewEsaClient("test-team", endpoint, "apikey", time.Duration(5))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s", r.URL.String())
	fmt.Fprintf(w, "hello")
}

func TestEsaGetPostsByUsername(t *testing.T) {
	h := map[string]http.Handler{
		"/v1/teams/test-team/posts": http.HandlerFunc(handler),
	}
	ts := NewTestEsaServer(t, h)
	c := NewTestEsaClient(t, ts.URL)

	res, err := c.GetPostsByUsername("achiku", time.Now(), time.Now(), 0)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", res)
}
