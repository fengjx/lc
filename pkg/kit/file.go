package kit

import (
	"os"
)

// IsFileOrDirExist 判断文件或目录是否存在
func IsFileOrDirExist(f string) (bool, error) {
	_, err := os.Stat(f)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
