package common

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/google/go-github/v61/github"

	"github.com/fengjx/lc/pkg/gh"
)

const versionPrefix = "v1"

// SyncTemplate 同步模板
// LatestDir 最新版本目录
func SyncTemplate(ctx context.Context) (LatestDir string, err error) {
	return downloadTemplate(ctx)
}

// 下载模板
func downloadTemplate(ctx context.Context) (LatestDir string, err error) {
	getDistFile := func(asset *github.ReleaseAsset) string {
		idStr := strconv.FormatInt(asset.GetID(), 10)
		LatestDir = getDownloadDir("template", idStr, "template.zip")
		return LatestDir
	}
	assertFilter := func(releases []*github.RepositoryRelease) *github.ReleaseAsset {
		sort.Slice(releases, func(i, j int) bool {
			return releases[i].PublishedAt.After(releases[j].PublishedAt.Time)
		})
		if len(releases) == 0 {
			return nil
		}
		release := releases[0]
		for _, item := range release.Assets {
			if item.Name != nil && *item.Name == "template.zip" {
				return item
			}
		}
		return nil
	}
	_, err = gh.DownloadReleaseAsset(ctx, GithubOwner, GithubRepo, assertFilter, getDistFile)

	return "", err
}

func getDownloadDir(paths ...string) string {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = "./"
	}
	ps := []string{homeDir, ".lc"}
	ps = append(ps, paths...)
	return filepath.Join(ps...)
}
