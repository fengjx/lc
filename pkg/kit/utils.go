package kit

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func IsIDEmpty(id interface{}) bool {
	idStr := ToString(id)
	return idStr == "" || idStr == "0"
}

func ToString(src interface{}) string {
	if src == nil {
		return ""
	}

	switch v := src.(type) {
	case string:
		return src.(string)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", src)
	case float32, float64:
		bs, _ := json.Marshal(v)
		return string(bs)
	case bool:
		if b, ok := src.(bool); ok && b {
			return "true"
		} else {
			return "false"
		}
	default:
		return fmt.Sprintf("%v", v)
	}
}

func ContainsString(collection []string, element string) bool {
	for _, item := range collection {
		if item == element {
			return true
		}
	}
	return false
}

// IsLastIndex 是否是最后一个元素
func IsLastIndex(idx int, s any) bool {
	return idx == reflect.ValueOf(s).Len()-1
}

func Add(y, x int) int {
	return x + y
}

func Sub(y, x int) int {
	return x - y
}
