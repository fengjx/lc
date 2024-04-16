package common

import (
	"embed"
	"encoding/json"
)

//go:embed .git-info.json
var embedFS embed.FS

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
	jsonBys, err := embedFS.ReadFile(".git-info.json")
	if err != nil {
		return info
	}
	_ = json.Unmarshal(jsonBys, &info)
	return info
}

func init() {
	GitInfo = getGitInfo()
}
