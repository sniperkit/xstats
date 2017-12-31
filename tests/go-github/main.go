package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	// "time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/roscopecoltran/sniperkit-agent/utils"

	"github.com/sniperkit/cacher"                               // cacher interface for http transporter
	cacher_badger "github.com/sniperkit/cacher/backends/badger" // BadgerKV default implementation

	"github.com/segmentio/stats/httpstats"
)

// var hc *http.Client

func main() {

	http.DefaultClient.Transport = httpstats.NewTransport(http.DefaultClient.Transport)

	cacheStoragePrefixPath := filepath.Join("shared", "cacher.badger")
	utils.EnsureDir(cacheStoragePrefixPath)
	hcache, err := cacher_badger.New(
		&cacher_badger.Config{
			ValueDir:    "api.github.com.v3.gzip", //gzip",
			StoragePath: cacheStoragePrefixPath,
			SyncWrites:  true,
			Debug:       false,
			Compress:    true,
		})
	if err != nil {
		panic(err)
	}

	t := cacher.NewTransport(hcache)
	t.MarkCachedResponses = true
	t.Debug = true
	t.Transport = httpstats.NewTransport(&http.Transport{})

	/*httpstats.NewTransport(
	&http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	})
	*/
	// timeout := time.Duration(10 * time.Second)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "aaed13d6a8b91302846dac0c8364dbfcecacd554"},
	)

	hc := &http.Client{
		Transport: &oauth2.Transport{
			Base:   t, // httpstats.NewTransport(t),
			Source: ts,
		},
		// Timeout: timeout,
	}

	gh := github.NewClient(hc)

	ctx := context.Background()

	var u string = ""
	var r string = ""

	if len(os.Args) > 2 {
		u = os.Args[1]
		r = os.Args[2]
	} else if len(os.Args) > 0 {
		log.Fatal("User and Repository name not specified!")
	}

	var opts = github.ListOptions{
		Page:    0,
		PerPage: 1000, /* This option is intended to just set the max of GitHub */
	}

	var maxPage int = 1

	for opts.Page <= maxPage {

		stars, response, e := gh.Activity.ListStargazers(ctx, u, r, &opts)

		log.Println("response.Header.Get(cacher.XFromCache)? ", response.Header.Get(cacher.XFromCache))

		if e != nil {
			log.Fatal(e)
			os.Exit(1)
		}

		for _, star := range stars {
			user := star.GetUser()
			time := star.GetStarredAt()

			fmt.Println(strconv.FormatInt(time.Time.Unix(), 10) + "\t" + user.GetLogin())
		}

		if opts.Page == maxPage {
			break
		}

		maxPage = response.LastPage
		opts.Page = response.NextPage
	}

}
