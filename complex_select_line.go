package extractor2

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/CreditTone/extractor2/html"
	"github.com/CreditTone/extractor2/json"
	estring "github.com/CreditTone/extractor2/string"
)

type SelectItem struct {
	InputType  string
	SelectBody string
}

type ComplexSelectLine struct {
	SelectItems []*SelectItem
	ResultType  []string
}

func NewSelectItem(selectitemStr string) *SelectItem {
	ret := SelectItem{}
	var subStart int
	if strings.HasPrefix(selectitemStr, "string(") {
		ret.InputType = "string"
		subStart = 7
	} else if strings.HasPrefix(selectitemStr, "json(") {
		ret.InputType = "json"
		subStart = 5
	} else if strings.HasPrefix(selectitemStr, "html(") {
		ret.InputType = "html"
		subStart = 5
	}
	ret.SelectBody = selectitemStr[subStart : len(selectitemStr)-1]
	return &ret
}

func NewComplexSelectLine(complexParseLine string) *ComplexSelectLine {
	var ret ComplexSelectLine
	complexParseLineAndTypes := strings.Split(complexParseLine, ">")
	complexParseLine = strings.TrimSpace(complexParseLineAndTypes[0])
	for i := 1; i < len(complexParseLineAndTypes); i++ {
		ret.ResultType = append(ret.ResultType, strings.TrimSpace(complexParseLineAndTypes[i]))
	}
	var selectitemStrs []string
	reg := regexp.MustCompile("\\s*(string|json|html)\\(")
	regResult := reg.FindAllStringSubmatch(complexParseLine, 100)
	var endWith int
	for i := 0; i < len(regResult); i++ {
		complexParseLine = strings.TrimSpace(complexParseLine[endWith:])
		if i == len(regResult)-1 {
			endWith = len(complexParseLine)
		} else {
			endWith = strings.Index(complexParseLine, regResult[i+1][0])
		}
		selectitemStrs = append(selectitemStrs, complexParseLine[0:endWith])
	}
	//fmt.Println(selectitemStrs...)
	for _, v := range selectitemStrs {
		ret.SelectItems = append(ret.SelectItems, NewSelectItem(v))
	}
	return &ret
}

func (self *ComplexSelectLine) doComplexSelectLineArray(body string) []interface{} {
	var lastSelectItem *SelectItem
	var previousSelectItem []*SelectItem
	if len(self.SelectItems) > 1 {
		previousSelectItem = append(previousSelectItem, self.SelectItems[0:len(self.SelectItems)-1]...)
	}
	lastSelectItem = self.SelectItems[len(self.SelectItems)-1]
	self.SelectItems = previousSelectItem
	body = self.doComplexSelectLineFirst(body)
	if lastSelectItem.InputType == "html" {
		return html.DoHtmlExtractor(lastSelectItem.SelectBody, body)
	} else if lastSelectItem.InputType == "json" {
		return json.DoJsonExtractor(lastSelectItem.SelectBody, body)
	} else if lastSelectItem.InputType == "string" {
		return estring.DoStringExtractor(lastSelectItem.SelectBody, body)
	}
	return nil
}

func (self *ComplexSelectLine) doComplexSelectLineFirst(body string) string {
	currentRet := body
	for _, selectItem := range self.SelectItems {
		if selectItem.InputType == "html" {
			currentRet = html.DoHtmlOneExtractor(selectItem.SelectBody, currentRet)
		} else if selectItem.InputType == "json" {
			jsonRet := json.DoJsonOneExtractor(selectItem.SelectBody, currentRet)
			currentRet = Interface2String(jsonRet)
		} else if selectItem.InputType == "string" {
			currentRet = estring.DoStringOneExtractor(selectItem.SelectBody, currentRet)
		}
	}
	return currentRet
}

func (self *ComplexSelectLine) doComplexSelectLine(body string) interface{} {
	currentRet := body
	for i, selectItem := range self.SelectItems {
		if selectItem.InputType == "html" {
			currentRet = html.DoHtmlOneExtractor(selectItem.SelectBody, currentRet)
		} else if selectItem.InputType == "json" {
			jsonRet := json.DoJsonOneExtractor(selectItem.SelectBody, currentRet)
			if jsonRet != nil && (i != len(self.SelectItems)-1 || len(self.ResultType) > 0) {
				currentRet = Interface2String(jsonRet)
			} else {
				return jsonRet
			}
		} else if selectItem.InputType == "string" {
			currentRet = estring.DoStringOneExtractor(selectItem.SelectBody, currentRet)
		}
	}
	var finalResult interface{}
	finalResult = currentRet
	for _, toType := range self.ResultType {
		finalResult = convertType(toType, currentRet)
		currentRet = fmt.Sprintf("%v", finalResult)
	}
	return finalResult
}
