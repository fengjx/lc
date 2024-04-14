package common

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/google/go-github/v61/github"
	"github.com/samber/lo"

	"github.com/fengjx/lc/pkg/gh"
	"github.com/fengjx/lc/pkg/kit"
)

const versionPrefix = "template-v1"

// SyncTemplate 同步模板
// LatestDir 最新版本目录
func SyncTemplate(ctx context.Context) (unzipDir string, err error) {
	downloadFile, err := downloadTemplate(ctx)
	if err != nil {
		return "", err
	}
	unzipDir = filepath.Join(filepath.Dir(downloadFile), "target")
	if exist, _ := kit.IsFileOrDirExist(unzipDir); exist {
		err = os.RemoveAll(unzipDir)
		if err != nil {
			return
		}
	}
	err = kit.Unzip(downloadFile, unzipDir)
	return
}

// 下载模板
func downloadTemplate(ctx context.Context) (downloadFile string, err error) {
	getDistFile := func(asset *github.ReleaseAsset, release *github.RepositoryRelease) string {
		downloadFile = getDownloadDir("res", release.GetTagName(), "template.zip")
		return downloadFile
	}
	assertFilter := func(releases []*github.RepositoryRelease) (*github.ReleaseAsset, *github.RepositoryRelease) {
		releases = lo.Filter(releases, func(item *github.RepositoryRelease, _ int) bool {
			//// 过滤 rc 版本
			if !useRC() && strings.Index(item.GetTagName(), "-rc") > 0 {
				return false
			}
			// 只兼容 template-v1-* 版本
			if !strings.HasPrefix(item.GetTagName(), versionPrefix) {
				return false
			}
			return true
		})
		if len(releases) == 0 {
			return nil, nil
		}
		sort.Slice(releases, func(i, j int) bool {
			return releases[i].PublishedAt.After(releases[j].PublishedAt.Time)
		})
		release := releases[0]
		for _, item := range release.Assets {
			if item.Name != nil && *item.Name == "template.zip" {
				return item, release
			}
		}
		return nil, nil
	}
	assert, exist, err := gh.DownloadReleaseAsset(ctx, GithubOwner, GithubRepo, assertFilter, getDistFile)
	if err == nil && !exist {
		color.Green("模板文件[%s]下载完成", assert.GetBrowserDownloadURL())
		color.Green("下载路径：%s", downloadFile)
	}
	return
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

func useRC() bool {
	use := os.Getenv("USE_RC")
	return lo.Contains([]string{"true", "1"}, strings.ToLower(use))
}
