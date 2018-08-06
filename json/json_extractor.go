package json

import (
	encodingJson "encoding/json"
	"strconv"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/elgs/jsonql"
	"github.com/xlvector/dlog"
)

const (
	ARRAY_DEFINE = "_array"
	FIRST_DEFINE = "_first"
)

func isEmptyParse(parseConfig map[string]interface{}) bool {
	for k, _ := range parseConfig {
		if !strings.HasPrefix(k, "_") {
			return false
		}
	}
	return true
}

//解析复杂结构
func DoJsonExtractor(array_define, val string) []interface{} {
	val = FilterJSONP(val)
	jsonData, err := simplejson.NewFromReader(strings.NewReader(val))
	if err != nil {
		dlog.Warn("读取json失败%s", err.Error())
		return nil
	}
	rootSelection := getJsonPaths(strings.Split(array_define, "."), jsonData)
	if rootSelection == nil {
		return nil
	}
	arraySelection := getJsonArray(rootSelection)
	var ret []interface{}
	for _, v := range arraySelection {
		ret = append(ret, v.Interface())
	}
	return ret
}

//解析单个结果
func DoJsonOneExtractor(parseConfig string, val string) interface{} {
	val = FilterJSONP(val)
	jsonData, err := simplejson.NewFromReader(strings.NewReader(val))
	if err != nil {
		dlog.Warn("读取json失败%s", err.Error())
		return nil
	}
	return queryValue(parseConfig, jsonData)
}

func doJsonExtractorSelection(parseConfig map[string]interface{}, jsonData *simplejson.Json) interface{} {
	array_define, array_define_ok := parseConfig[ARRAY_DEFINE]
	first_define, first_define_ok := parseConfig[FIRST_DEFINE]
	rootSelection := jsonData
	if first_define_ok {
		rootSelection = getJsonPaths(strings.Split(first_define.(string), "."), rootSelection)
	}
	if array_define_ok {
		rootSelection = getJsonPaths(strings.Split(array_define.(string), "."), rootSelection)
		arraySelection := getJsonArray(rootSelection)
		isConfigEmpty := isEmptyParse(parseConfig)
		arrayRet := []interface{}{}
		for _, itemSelection := range arraySelection {
			if isConfigEmpty {
				arrayRet = append(arrayRet, itemSelection.Interface())
			} else {
				item := map[string]interface{}{}
				for key, parseVal := range parseConfig {
					if strings.HasPrefix(key, "_") {
						continue
					}
					if parseValQuery, ok := parseVal.(string); ok {
						item[key] = queryValue(parseValQuery, itemSelection)
					}
					if parseMapQuery, ok := parseVal.(map[string]interface{}); ok {
						item[key] = doJsonExtractorSelection(parseMapQuery, itemSelection)
					}
				}
				arrayRet = append(arrayRet, item)
			}
		}
		return arrayRet
	}
	mapRet := map[string]interface{}{}
	for key, parseVal := range parseConfig {
		if strings.HasPrefix(key, "_") {
			continue
		}
		if parseValQuery, ok := parseVal.(string); ok {
			mapRet[key] = queryValue(parseValQuery, rootSelection)
		}
		if parseMapQuery, ok := parseVal.(map[string]interface{}); ok {
			mapRet[key] = doJsonExtractorSelection(parseMapQuery, rootSelection)
		}
	}
	return mapRet

}

func getJsonArray(json *simplejson.Json) []*simplejson.Json {
	arrInterface, err := json.Array()
	if err == nil {
		jsonArr := []*simplejson.Json{}
		for index, _ := range arrInterface {
			item := &simplejson.Json{}
			item.SetPath([]string{}, arrInterface[index])
			jsonArr = append(jsonArr, item)
		}
		return jsonArr
	} else {
		dlog.Warn("转换jsonarray失败 %v", err)
	}
	return nil
}

func getJsonArrayLength(json *simplejson.Json) int {
	arrInterface, err := json.Array()
	if err == nil {
		return len(arrInterface)
	} else {
		dlog.Warn("转换jsonarray失败 %v", err)
	}
	return -1
}

func getJsonPaths(jsonpath []string, json *simplejson.Json) *simplejson.Json {
	for x := 0; x < len(jsonpath); x++ {
		if json == nil {
			break
		}
		cmd := jsonpath[x]
		if strings.HasPrefix(cmd, "[") && strings.HasSuffix(cmd, "]") {
			indexStr := cmd[1 : len(cmd)-1]
			if strings.Contains(indexStr, ":") {
				arrInterface, err := json.Array()
				if err != nil {
					dlog.Warn("转换jsonarray失败 %v", err)
					return nil
				}
				var startIndex int
				var endIndex int
				startIndexStr := strings.Split(indexStr, ":")[0]
				endIndexStr := strings.Split(indexStr, ":")[1]
				if len(startIndexStr) > 0 {
					index, err := strconv.Atoi(startIndexStr)
					if err != nil {
						dlog.Warn("convet int error %v", err)
						return nil
					}
					startIndex = index
				} else {
					startIndex = 0
				}
				if len(endIndexStr) > 0 {
					index, err := strconv.Atoi(endIndexStr)
					if err != nil {
						dlog.Warn("convet int error %v", err)
						return nil
					}
					endIndex = index
				} else {
					endIndex = len(arrInterface)
				}
				arrJson := &simplejson.Json{}
				arrJson.SetPath([]string{}, arrInterface[startIndex:endIndex])
				return arrJson
			} else {
				index, err := strconv.Atoi(indexStr)
				if err != nil {
					dlog.Warn("convet int error %v", err)
					json = nil
				}
				temp := json.GetIndex(index)
				if temp != nil {
					json = temp
				} else {
					dlog.Warn("json array not found %d", index)
					json = nil
				}
			}
		} else if strings.HasPrefix(cmd, "(") && strings.HasSuffix(cmd, ")") {
			query := cmd[1 : len(cmd)-1]
			parser := jsonql.NewQuery(json.Interface())
			m, err := parser.Query(query)
			if err != nil {
				dlog.Warn("not found %s", err.Error())
				return json
			}
			v, err := encodingJson.Marshal(m)
			if err != nil {
				dlog.Warn("Marshal json %s", err.Error())
				return json
			}
			newjson, err := simplejson.NewJson(v)
			if err != nil {
				dlog.Warn("NewJson %s", err.Error())
				return json
			}
			json = newjson
		} else {
			temp, exist := json.CheckGet(cmd)
			if !exist {
				yaunshi, _ := encodingJson.Marshal(json.Interface())
				dlog.Info("json不存在key %s from %s", cmd, string(yaunshi))
				json = nil
			} else {
				json = temp
			}
		}
	}
	return json
}

type JsonSelector struct {
	JsonPath string
}

func NewJsonSelector(v string) *JsonSelector {
	ret := &JsonSelector{}
	tks := strings.Split(v, ";")
	ret.JsonPath = tks[0]
	return ret
}

func queryValue(jsonPath string, json *simplejson.Json) interface{} {
	defer func() {
		if err := recover(); err != nil {
			dlog.Warn("json解析错误 %v", err)
		}
	}()
	var ret interface{}
	jsonSelector := NewJsonSelector(jsonPath)
	if len(jsonSelector.JsonPath) > 0 {
		b := getJsonPaths(strings.Split(jsonSelector.JsonPath, "."), json)
		if b != nil {
			ret = b.Interface()
		}
	} else {
		ret = json.Interface()
	}
	return ret
}
