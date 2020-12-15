package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"time"
)

var (
	apikey = flag.String("apikey", "", "esa api key")
	user   = flag.String("user", "", "esa username")
	team   = flag.String("team", "", "esa team")
)

func main() {
	flag.Parse()

	c, err := NewEsaClient(*team, *apikey, time.Duration(5))
	if err != nil {
		log.Fatal(err)
	}
	res, err := c.GetPostsByUsername(*user, time.Now(), time.Now(), 1)
	if err != nil {
		log.Fatal(err)
	}
	posts := res.Posts
	nextPage := res.NextPage
	for nextPage != 0 {
		res, err := c.GetPostsByUsername(*user, time.Now(), time.Now(), nextPage)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, res.Posts...)
		nextPage = res.NextPage
		time.Sleep(1)

	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	for _, r := range posts {
		fmt.Printf("- [%s](%s) %s\n", r.FullName, r.URL, r.CreatedAt.Format("2006/1/2 15:04"))
	}
}
