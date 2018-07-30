package string

import (
	"regexp"

	"github.com/xlvector/dlog"
)

func DoStringExtractor(array_define, val string) []interface{} {
	defer func() {
		if err := recover(); err != nil {
			dlog.Warn("json解析错误%v", err)
		}
	}()
	var ret []interface{}
	reg := regexp.MustCompile(array_define)
	result := reg.FindAllStringSubmatch(val, 1000000)
	for _, group := range result {
		if len(group) > 1 {
			ret = append(ret, group[1])
		} else {
			ret = append(ret, group[0])
		}
	}
	return ret
}

//解析单个结果
func DoStringOneExtractor(parseConfig string, val string) string {
	reg := regexp.MustCompile(parseConfig)
	result := reg.FindAllStringSubmatch(val, 1)
	if len(result) > 0 {
		group := result[0]
		if len(group) > 1 {
			return group[1]
		} else {
			return group[0]
		}
	} else {
		dlog.Warn("正则%s没有匹配到结果", parseConfig)
	}
	return ""
}
