package main

import (
	"context"

	"github.com/google/go-github/v38/github"
	"golang.org/x/oauth2"
)

func newGithubClient(accessToken string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func isGithubPrivate(client *github.Client, user string, repoName string) (bool, error) {
	repo, _, err := client.Repositories.Get(context.Background(), user, repoName)

	if err != nil {
		return false, err
	}

	return *repo.Private, nil
}
