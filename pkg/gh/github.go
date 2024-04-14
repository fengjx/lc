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

type AssetFilter func([]*github.RepositoryRelease) (asset *github.ReleaseAsset, release *github.RepositoryRelease)

// DownloadReleaseAsset 下载 release 资源
func DownloadReleaseAsset(ctx context.Context, owner, repo string, filter AssetFilter, getDistFile func(asset *github.ReleaseAsset, release *github.RepositoryRelease) string) (asset *github.ReleaseAsset, exist bool, err error) {
	releases, _, err := cli.Repositories.ListReleases(ctx, owner, repo, nil)
	if err != nil {
		return nil, false, err
	}
	if len(releases) == 0 {
		return nil, false, err
	}
	asset, release := filter(releases)
	if asset == nil {
		return nil, false, fmt.Errorf("no release assert found from https://github.com/%s/%s/releases", owner, repo)
	}
	distFile := getDistFile(asset, release)
	if ok, _ := kit.IsFileOrDirExist(distFile); ok {
		return asset, true, nil
	}
	rc, _, err := cli.Repositories.DownloadReleaseAsset(ctx, owner, repo, *asset.ID, httpcli.GetClient())
	if err != nil {
		return nil, false, err
	}
	defer rc.Close()
	_, err = kit.Copy(distFile, rc)
	if err != nil {
		return nil, false, err
	}
	return asset, false, nil
}
