package extractor2

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var typeFunc map[string]func(interface{}) interface{}

func init() {
	typeFunc = make(map[string]func(interface{}) interface{})
	typeFunc["int"] = toInt
	typeFunc["float"] = toFloat
	typeFunc["boolean"] = toBoolean
	typeFunc["htmlText"] = toHtmlText
	typeFunc["string"] = toString
	typeFunc["unicodeText"] = unicodeText
}

func convertType(ty string, val interface{}) interface{} {
	convertFunc := typeFunc[ty]
	if convertFunc != nil && val != nil {
		return convertFunc(val)
	}
	return nil
}

func unicodeText(val interface{}) interface{} {
	str := fmt.Sprintf("%v", val)
	reg := regexp.MustCompile("[^\\w]{1}[uU]{1}([0-9A-Fa-f]{4})")
	result := reg.FindAllStringSubmatch(str, 1000000)
	for _, group := range result {
		old := group[0]
		if strings.HasPrefix(old, "\\u") {
			v := group[1]
			temp, err := strconv.ParseInt(v, 16, 32)
			if err != nil {
				continue
			}
			unicodeChar := fmt.Sprintf("%c", temp)
			str = strings.Replace(str, old, unicodeChar, -1)
		}
	}
	return str
}

func toInt(val interface{}) interface{} {
	str := fmt.Sprintf("%v", val)
	if strings.Contains(str, "e") || strings.Contains(str, "E") {
		var newf float64
		n, err := fmt.Sscanf(str, "%e", &newf)
		if err != nil {
			fmt.Println("取科学计数发生错误 ", err.Error())
			return nil
		} else if 1 != n {
			fmt.Println("n is not one")
			return nil
		}
		return int64(newf)
	}
	intVal, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println("不能将%v转换成int", val)
		return nil
	}
	return intVal
}

func toFloat(val interface{}) interface{} {
	str := fmt.Sprintf("%v", val)
	if strings.Contains(str, "e") || strings.Contains(str, "E") {
		var newf float64
		n, err := fmt.Sscanf(str, "%e", &newf)
		if err != nil {
			fmt.Errorf("取科学计数发生错误%v", err.Error())
			return nil
		} else if 1 != n {
			fmt.Errorf("n is not one")
			return nil
		}
		return newf
	}
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fmt.Println("不能将%v转换成float", val, err)
		return nil
	}
	return f
}

func toBoolean(val interface{}) interface{} {
	str := fmt.Sprintf("%v", val)
	if strings.ToLower(str) == "true" {
		return true
	}
	if strings.ToLower(str) == "false" {
		return false
	}
	return nil
}

func toString(val interface{}) interface{} {
	return fmt.Sprintf("%v", val)
}

func toHtmlText(val interface{}) interface{} {
	str := unicodeText(val).(string)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(str))
	if err != nil {
		fmt.Println("读取html失败 %v", err.Error())
		return nil
	}
	return strings.TrimSpace(doc.Text())
}
