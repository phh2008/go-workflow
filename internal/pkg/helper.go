package pkg

import (
	"bytes"
	"encoding/json"
)

// JSONToStruct 将 JSON 字符串反序列化为指定的结构体
func JSONToStruct(j string, s any) error {
	return json.Unmarshal([]byte(j), s)
}

// JSONMarshal 将任意值序列化为 JSON 字节切片。
// escapeHTML 控制是否对 HTML 特殊字符进行转义，
// 当 escapeHTML 为 true 时，<、>、&、U+2028、U+2029 会被转义。
func JSONMarshal(t any, escapeHTML bool) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(escapeHTML)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

// MakeUnique 利用 Map 对多个字符串切片进行去重处理，返回去重后的结果切片
func MakeUnique(lists ...[]string) []string {
	set := make(map[string]string)

	for _, list := range lists {
		for _, item := range list {
			set[item] = ""
		}
	}

	unique := make([]string, 0, len(set))

	for k := range set {
		unique = append(unique, k)
	}

	return unique
}

// RemoveAllElements 删除字符串切片中所有与指定值相等的元素，返回新切片
func RemoveAllElements(slice []string, value string) []string {
	result := []string{}
	for _, v := range slice {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}
