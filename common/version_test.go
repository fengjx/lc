package common_test

import (
	"testing"

	"github.com/fengjx/lc/common"
)

func TestVersion_GetGitInfo(t *testing.T) {
	gitInfo := common.GitInfo
	t.Log(gitInfo.Branch, gitInfo.Hash)
}
