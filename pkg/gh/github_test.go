package gh_test

import (
	"testing"

	"github.com/fengjx/lc/pkg/gh"
)

func TestGetReleaseInfo(t *testing.T) {
	info, err := gh.GetReleaseInfo("fengjx", "luchen")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}
