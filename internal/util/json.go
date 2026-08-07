package util

import "encoding/json"

func JsonEncode(data any) []byte {
	jsonData, _ := json.Marshal(data)

	return jsonData
}

func JsonEncodeStr(data any) string {
	return string(JsonEncode(data))
}
