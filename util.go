package extractor2

import (
	"strings"
	"zhongguo/extractor2/json"
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
