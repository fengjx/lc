package filegen

import (
	"bytes"
	"encoding/json"
	"go/format"
	"strings"
)

type formater func([]byte) ([]byte, error)

func goFormater(src []byte) ([]byte, error) {
	return format.Source(src)
}

func jsonFormater(src []byte) ([]byte, error) {
	// 格式化 JSON
	var formatted bytes.Buffer
	err := json.Indent(&formatted, src, "", "  ")
	if err != nil {
		return nil, err
	}
	return formatted.Bytes(), nil
}

func getFormater(targetFile string) formater {
	if strings.HasSuffix(targetFile, ".go") {
		return goFormater
	} else if strings.HasSuffix(targetFile, ".json") {
		return jsonFormater
	}
	return nil
}
