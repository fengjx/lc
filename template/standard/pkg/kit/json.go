package kit

import "github.com/fengjx/go-halo/json"

// MustToJSON 对象转 json 字符串，忽略异常
func MustToJSON(data any) string {
	jsonStr, _ := json.ToJson(data)
	return jsonStr
}
