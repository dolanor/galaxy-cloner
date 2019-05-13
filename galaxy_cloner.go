package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	var (
		githubToken = os.Getenv("GALAXY_CLONER_GITHUB_TOKEN")
		repoPerPage = 100
		// Careful with that, you might get rate limited
		concurrencyEnv = os.Getenv("GALAXY_CLONER_CONCURRENCY")
		concurrency    int
		dstOrg         = os.Getenv("GALAXY_CLONER_DEST_ORG")
	)

	var err error
	if concurrencyEnv != "" {
		concurrency, err = strconv.Atoi(concurrencyEnv)
		if err != nil {
			panic("GALAXY_CLONER_concurrency misconfigured" + err.Error())
		}
	} else {
		concurrency = runtime.NumCPU()
	}

	if dstOrg == "" {
		panic("GALAXY_CLONER_DEST_ORG misconfigured")
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := github.NewClient(httpClient)

	ctx := context.Background()

	// We allocate enough space in the channel to avoid blocking the starred repo
	// gathering.
	toFork := make(chan *github.Repository, repoPerPage*2)
	for i := 0; i < concurrency; i++ {
		go repoForker(ctx, client, toFork, dstOrg)
	}

	var resp = &github.Response{}
	var opt = &github.ActivityListStarredOptions{
		ListOptions: github.ListOptions{
			PerPage: repoPerPage,
			Page:    5,
		},
	}

	// We make sure we enter the loop
	resp.NextPage = 1
	for resp.NextPage != 0 {
		starred, resp, err := client.Activity.ListStarred(ctx, "", opt)
		if err != nil {
			log.Println("can't list starred repositories:", err)
		}

		// We send the repos of this page to get forked
		for _, r := range starred {
			toFork <- r.Repository
		}

		if resp.NextPage == 0 {
			break
		}
		// We set to get the next page
		opt.Page = resp.NextPage
	}
}

// repoForker get repo from the in chan and fork the repository in the org if it doesn't exist already
func repoForker(ctx context.Context, client *github.Client, in <-chan *github.Repository, org string) {
	forkOpt := &github.RepositoryCreateForkOptions{
		Organization: org,
	}

	for repo := range in {
		fmt.Printf("\"%s/%s\": checking\n", *repo.Owner.Login, *repo.Name)
		rep, resp, err := client.Repositories.Get(ctx, org, *repo.Name)
		if err != nil {
			if e, ok := err.(*github.ErrorResponse); ok && e.Response.StatusCode != 404 {
				fmt.Println("unhandled error:", err)
				continue
			}
		}

		// We can find a repository, so we won't fork
		// TODO update the fork to the latest upstream version
		if resp.Response.StatusCode >= 200 && resp.Response.StatusCode < 300 {
			fmt.Printf("\"%s/%s\": already cloned, let's skip this one.\n", *repo.Owner.Login, *repo.Name)
			continue
		}
		fmt.Printf("\"%s/%s\": %d wait what?: %v\n", *repo.Owner.Login, *repo.Name, resp.Response.StatusCode, err)

		// We create a fork for the starred repository
		rep, _, err = client.Repositories.CreateFork(ctx, *repo.Owner.Login, *repo.Name, forkOpt)
		if err != nil {
			switch err.(type) {
			case *github.AcceptedError:
				fmt.Printf("\"%s/%s\" fork queued in GitHub: %v\n", *rep.Owner.Login, *rep.Name, err)
			default:
				log.Println("couldn't fork:", err)
			}
		}
	}
}
