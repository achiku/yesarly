package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"testing"
	"time"
)

func TestGetPostsByUsername(t *testing.T) {
	apikey := os.Getenv("ESA_API_KEY")
	c, err := NewEsaClient("kanmu", apikey, time.Duration(5))
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
