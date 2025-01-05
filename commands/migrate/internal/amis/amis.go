package amis

import "github.com/fengjx/lc/commands/migrate/internal/types"

const DefaultInputType = "input-text"

var InputTypeMap = map[string]string{
	types.Date:          "input-datetime",
	types.DateTime:      "input-datetime",
	types.SmallDateTime: "input-datetime",
	types.Time:          "input-datetime",
	types.TimeStamp:     "input-datetime",
	types.TimeStampz:    "input-datetime",
	types.Year:          "input-datetime",
}

// InputType sqlType 映射 input type
func InputType(sqlType string) string {
	if input, ok := InputTypeMap[sqlType]; ok {
		return input
	}
	return DefaultInputType
}
