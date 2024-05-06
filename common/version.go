package common

import (
	"embed"
	"encoding/json"
	"strings"
)

//go:embed .git-info.json
var embedFS embed.FS

const (
	GithubOwner = "fengjx"
	GithubRepo  = "lc"
	v           = "v1.0.0"
)

var GitInfo gitInfo

type gitInfo struct {
	Branch  string `json:"branch"`
	Hash    string `json:"hash"`
	Version string `json:"version"`
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

func (g gitInfo) GetVersion() string {
	if g.Version != "" {
		return g.Version
	}
	return strings.Join([]string{v, g.Branch, g.Hash}, "-")
}
