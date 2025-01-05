package gh_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-github/v61/github"

	"github.com/fengjx/lc/pkg/gh"
)

func TestGetReleaseI(t *testing.T) {
	ctx := context.Background()
	info, err := gh.GetRelease(ctx, "fengjx", "lc", func(releases []*github.RepositoryRelease) *github.RepositoryRelease {
		return releases[0]
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}

func TestDownloadReleaseAsset(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	getDownloadDir := func(asset *github.ReleaseAsset, release *github.RepositoryRelease) string {
		return ".lc/template/start-template.zip"
	}
	assetFilter := func(releases []*github.RepositoryRelease) (*github.ReleaseAsset, *github.RepositoryRelease) {
		sort.Slice(releases, func(i, j int) bool {
			return releases[i].PublishedAt.After(releases[j].PublishedAt.Time)
		})
		if len(releases) == 0 {
			return nil, nil
		}
		release := releases[0]
		for _, item := range release.Assets {
			if item.Name != nil && *item.Name == "start-template.zip" {
				return item, release
			}
		}
		return nil, nil
	}
	asset, _, err := gh.DownloadReleaseAsset(ctx, "fengjx", "lc", assetFilter, getDownloadDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("asset", *asset.BrowserDownloadURL)
}
