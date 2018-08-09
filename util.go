package extractor2

import (
	system_json "encoding/json"
	"fmt"
	"strings"

	"github.com/CreditTone/extractor2/json"
)

//0表示html 1表示json 2表示既不是html也不是json
func DetectContentType(content string) int {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "<") && strings.HasSuffix(content, ">") {
		return 0
	}
	if strings.Contains(content, "<html") || strings.Contains(content, "<body") || strings.Contains(content, "<a") || strings.Contains(content, "<p") || strings.Contains(content, "<span") || strings.Contains(content, "<div") {
		return 0
	}
	content = json.FilterJSONP(content)
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		return 1
	}
	if strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]") {
		return 1
	}
	return 2
}
func Interface2String(val interface{}) string {
	if str, ok := val.(string); ok {
		return str
	}
	if i, ok := val.(int); ok {
		return fmt.Sprintf("%v", i)
	}
	if f32, ok := val.(float32); ok {
		return fmt.Sprintf("%v", f32)
	}
	if f64, ok := val.(float64); ok {
		return fmt.Sprintf("%v", f64)
	}
	if b, ok := val.(bool); ok {
		return fmt.Sprintf("%v", b)
	}
	if r, ok := val.(rune); ok {
		return fmt.Sprintf("%v", r)
	}
	if bs, ok := val.([]byte); ok {
		return fmt.Sprintf("%v", bs)
	}
	if marshalJson, err := system_json.Marshal(val); err == nil {
		return string(marshalJson)
	}
	return fmt.Sprintf("%v", val)
}
