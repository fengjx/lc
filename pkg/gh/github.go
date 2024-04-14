package gh

import (
	"context"
	"fmt"

	"github.com/google/go-github/v61/github"

	"github.com/fengjx/lc/pkg/httpcli"
	"github.com/fengjx/lc/pkg/kit"
)

var cli = github.NewClient(nil)

type ReleaseFilter func([]*github.RepositoryRelease) *github.RepositoryRelease

// GetRelease 查询 github release
func GetRelease(ctx context.Context, owner, repo string, filter ReleaseFilter) (*github.RepositoryRelease, error) {
	releases, _, err := cli.Repositories.ListReleases(ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}
	return filter(releases), nil
}

type AssetFilter func([]*github.RepositoryRelease) *github.ReleaseAsset

// DownloadReleaseAsset 下载 release 资源
func DownloadReleaseAsset(ctx context.Context, owner, repo string, filter AssetFilter, getDistFile func(asset *github.ReleaseAsset) string) (asset *github.ReleaseAsset, err error) {
	releases, _, err := cli.Repositories.ListReleases(ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}
	if len(releases) == 0 {
		return nil, err
	}
	asset = filter(releases)
	if asset == nil {
		return nil, fmt.Errorf("no release assert found from https://github.com/%s/%s/releases", owner, repo)
	}
	rc, _, err := cli.Repositories.DownloadReleaseAsset(ctx, owner, repo, *asset.ID, httpcli.GetClient())
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	_, err = kit.Copy(getDistFile(asset), rc)
	if err != nil {
		return nil, err
	}
	return asset, nil
}
