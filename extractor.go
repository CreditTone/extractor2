package extractor2

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/CreditTone/extractor2/context"
)

const (
	ARRAY_DEFINE = "_array"
	FIRST_DEFINE = "_first"
)

func init() {
	//	json.ExternalDoHtmlOneExtractor = html.DoHtmlOneExtractor
	//	json.ExternalDoStringOneExtractor = estring.DoStringOneExtractor
	//	html.ExternalDoJsonOneExtractor = json.DoJsonOneExtractor
	//	html.ExternalDoStringOneExtractor = estring.DoStringOneExtractor
	//	estring.ExternalDoHtmlOneExtractor = html.DoHtmlOneExtractor
	//	estring.ExternalDoJsonOneExtractor = json.DoJsonOneExtractor
}

type Extractor struct {
	context *context.Context
}

func NewExtractor(doTemplateFunc func(template string) string) *Extractor {
	instance := Extractor{}
	if doTemplateFunc == nil {
		instance.context = context.NewContext(func(template string) string {
			return template
		})
	} else {
		instance.context = context.NewContext(doTemplateFunc)
	}
	return &instance
}

func isEmptyParse(parseConfig map[string]interface{}) bool {
	for k, _ := range parseConfig {
		if !strings.HasPrefix(k, "_") {
			return false
		}
	}
	return true
}

func (self *Extractor) convertArrayOfString(input []interface{}) []string {
	var ret []string
	for _, v := range input {
		ret = append(ret, Interface2String(v))
	}
	return ret
}

func (self *Extractor) Do(parseConfig map[string]interface{}, body string) interface{} {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("解析路径出错:%v", parseConfig)
			debug.PrintStack()
		}
	}()
	first_define, first_define_ok := parseConfig[FIRST_DEFINE]
	array_define, array_define_ok := parseConfig[ARRAY_DEFINE]
	if first_define_ok {
		body = self.doParseFirst(first_define.(string), body)
	}
	if array_define_ok {
		array_ret := self.doParseArray(array_define.(string), body)
		isConfigEmpty := isEmptyParse(parseConfig)
		if isConfigEmpty {
			return array_ret
		}
		arrayOfString := self.convertArrayOfString(array_ret)
		arrayRet := []interface{}{}
		for _, itemBody := range arrayOfString {
			item := map[string]interface{}{}
			for key, parseVal := range parseConfig {
				if strings.HasPrefix(key, "_") {
					continue
				}
				if parseValQuery, ok := parseVal.(string); ok {
					item[key] = self.doParseFinalResult(parseValQuery, itemBody)
				}
				if parseMapQuery, ok := parseVal.(map[string]interface{}); ok {
					item[key] = self.Do(parseMapQuery, itemBody)
				}
			}
			arrayRet = append(arrayRet, item)
		}
		return arrayRet
	}
	mapRet := map[string]interface{}{}
	for key, parseVal := range parseConfig {
		if strings.HasPrefix(key, "_") {
			continue
		}
		if parseValQuery, ok := parseVal.(string); ok {
			mapRet[key] = self.doParseFinalResult(parseValQuery, body)
		}
		if parseMapQuery, ok := parseVal.(map[string]interface{}); ok {
			mapRet[key] = self.Do(parseMapQuery, body)
		}
	}
	return mapRet
}

func (self *Extractor) DoOne(parseConfig string, body string) interface{} {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("解析路径出错:%s", parseConfig)
			debug.PrintStack()
		}
	}()
	return self.doParseFinalResult(parseConfig, body)
}

func (self *Extractor) doParseArray(complexParseLine string, body string) []interface{} {
	if strings.Contains(complexParseLine, "{{") && strings.Contains(complexParseLine, "}}") {
		complexParseLine = self.context.ParseTmplate(complexParseLine)
	}
	complexSelectLine := NewComplexSelectLine(complexParseLine)
	return complexSelectLine.doComplexSelectLineArray(body)
}

func (self *Extractor) doParseFirst(complexParseLine string, body string) string {
	if strings.HasPrefix(complexParseLine, "{{") && strings.HasSuffix(complexParseLine, "}}") {
		return self.context.ParseTmplate(complexParseLine)
	}
	if strings.Contains(complexParseLine, "{{") && strings.Contains(complexParseLine, "}}") {
		complexParseLine = self.context.ParseTmplate(complexParseLine)
	}
	complexSelectLine := NewComplexSelectLine(complexParseLine)
	return complexSelectLine.doComplexSelectLineFirst(body)
}

func (self *Extractor) doParseFinalResult(complexParseLine string, body string) interface{} {
	if strings.HasPrefix(complexParseLine, "{{") && strings.HasSuffix(complexParseLine, "}}") {
		return self.context.ParseTmplate(complexParseLine)
	}
	if strings.Contains(complexParseLine, "{{") && strings.Contains(complexParseLine, "}}") {
		complexParseLine = self.context.ParseTmplate(complexParseLine)
	}
	complexSelectLine := NewComplexSelectLine(complexParseLine)
	return complexSelectLine.doComplexSelectLine(body)
}
