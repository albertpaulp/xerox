package gitclient

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func getAccessToken() string {
	accessToken, err := os.LookupEnv("GITHUB_ACCESS_TOKEN")
	if err != true {
		fmt.Println("Error")
		panic("GITHUB_ACCESS_TOKEN is not found!")
	}
	return accessToken
}

// Client returns pointer to github.Client
func Client() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: getAccessToken()},
	)
	tc := oauth2.NewClient(ctx, ts)
	log.Println("Connected to Github")
	return github.NewClient(tc)
}
