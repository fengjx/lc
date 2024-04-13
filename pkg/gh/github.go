package gh

import (
	"context"

	"github.com/google/go-github/v61/github"
)

var cli = github.NewClient(nil)

type ReleaseFilter func([]*github.RepositoryRelease) *github.RepositoryRelease

func GetReleaseInfo(repo, proj string, filter ReleaseFilter) (*github.RepositoryRelease, error) {
	releases, _, err := cli.Repositories.ListReleases(context.Background(), repo, proj, nil)
	if err != nil {
		return nil, err
	}
	if len(releases) == 0 {
		return nil, nil
	}
	return filter(releases), nil
}
