package parser

import (
	"strconv"
)

type Uint64 uint64

func (id *Uint64) UnmarshalJSON(data []byte) error {
	// 移除 JSON 字符串中的引号
	strValue, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	// 将字符串解析为 uint64
	uintValue, err := strconv.ParseUint(strValue, 10, 64)
	if err != nil {
		return err
	}

	// 将 uint64 赋值给 ID
	*id = Uint64(uintValue)
	return nil
}
