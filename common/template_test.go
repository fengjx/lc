package common_test

import (
	"context"
	"testing"

	"github.com/fengjx/lc/common"
)

func TestSyncTemplate(t *testing.T) {
	ctx := context.Background()
	LatestDir, err := common.SyncTemplate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("LatestDir", LatestDir)
}
