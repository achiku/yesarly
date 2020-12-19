package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"time"
)

var (
	apikey = flag.String("apikey", "", "esa api key (required)")
	user   = flag.String("user", "", "esa username (required)")
	team   = flag.String("team", "", "esa team (required)")
	from   = flag.String("from", "", "from (format: 2020-01-01) (default: 1/1 of this year)")
	to     = flag.String("to", "", "to (format: 2020-12-31) (default: 12/31 of this year)")
)

func main() {
	flag.Parse()

	u := "https://api.esa.io/"
	c, err := NewEsaClient(*team, u, *apikey, time.Duration(20))
	if err != nil {
		log.Fatal(err)
	}

	if *apikey == "" {
		log.Fatal("apikey is required")
	}
	if *user == "" {
		log.Fatal("user is required")
	}
	if *team == "" {
		log.Fatal("team is required")
	}

	var pFrom time.Time
	if *from == "" {
		pFrom = time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.Local)
	} else {
		pFrom, err = time.Parse("2006-01-02", *from)
		if err != nil {
			log.Fatal(err)
		}
	}

	var pTo time.Time
	if *to == "" {
		pTo = time.Date(time.Now().Year(), 12, 31, 0, 0, 0, 0, time.Local)
	} else {
		pTo, err = time.Parse("2006-01-02", *to)
		if err != nil {
			log.Fatal(err)
		}
	}

	res, err := c.GetPostsByUsernameWithRetry(*user, pFrom, pTo, 1)
	if err != nil {
		log.Fatal(err)
	}

	numPosts := res.TotalCount
	var numPages int
	if res.TotalCount%20 == 0 {
		numPages = res.TotalCount / 20
	} else {
		numPages = (res.TotalCount / 20) + 1
	}
	log.Printf("total posts:%d, total page: %d", numPosts, numPages)
	log.Printf("current page %d", res.Page)

	posts := res.Posts
	nextPage := res.NextPage
	for nextPage != 0 {
		res, err := c.GetPostsByUsernameWithRetry(*user, pFrom, pTo, nextPage)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("current page %d", res.Page)
		posts = append(posts, res.Posts...)
		nextPage = res.NextPage
		time.Sleep(time.Duration(2) * time.Second)
	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	for _, r := range posts {
		fmt.Printf("- [%s](%s) %s\n", r.FullName, r.URL, r.CreatedAt.Format("2006/01/02 15:04"))
	}
}
