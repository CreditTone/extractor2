package string

import (
	"regexp"
	"strings"
	"zhongguo/extractor2/context"

	"github.com/xlvector/dlog"
)

func DoStringExtractor(parseConfig map[string]interface{}, val string, context *context.Context) interface{} {
	defer func() {
		if err := recover(); err != nil {
			dlog.Warn("json解析错误%v", err)
		}
	}()

	ret := make(map[string]interface{})
	for key, parseVal := range parseConfig {
		if strings.HasPrefix(key, "_") {
			continue
		}
		if parseValQuery, ok := parseVal.(string); ok {
			if strings.HasPrefix(parseValQuery, "{{") && strings.HasSuffix(parseValQuery, "}}") {
				ret[key] = context.ParseTmplate(parseValQuery)
				continue
			}
			if strings.Contains(parseValQuery, "{{") && strings.Contains(parseValQuery, "}}") {
				parseValQuery = context.ParseTmplate(parseValQuery)
			}
			reg := regexp.MustCompile(parseValQuery)
			result := reg.FindAllStringSubmatch(val, 1)
			if len(result) > 0 {
				group := result[0]
				if len(group) > 1 {
					ret[key] = group[1]
				} else {
					ret[key] = group[0]
				}
			} else {
				dlog.Warn("正则%s没有匹配到结果", parseValQuery)
			}
		}
		if parseMapQuery, ok := parseVal.(map[string]interface{}); ok {
			ret[key] = DoStringExtractor(parseMapQuery, val, context)
		}

	}
	return ret
}
