package common

import "github.com/fengjx/lc/pkg/kit"

const (
	GithubOwner = "fengjx"
	GithubRepo  = "lc"
)

var GitInfo gitInfo

type gitInfo struct {
	Branch string `json:"branch"`
	Hash   string `json:"hash"`
}

func getGitInfo() gitInfo {
	info := gitInfo{}
	gitInfoPath, _ := kit.Lookup(".git-info.json", 5)
	if gitInfoPath != "" {
		_ = kit.ReadJSONFile(gitInfoPath, &info)
	}
	return info
}

func init() {
	GitInfo = getGitInfo()
}
