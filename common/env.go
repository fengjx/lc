package common

import (
	"os"
	"strings"

	"github.com/samber/lo"
)

// IsDebug 是否是调试模式
func IsDebug() bool {
	debug := os.Getenv("DEBUG")
	return lo.Contains([]string{"true", "1"}, strings.ToLower(debug))
}
