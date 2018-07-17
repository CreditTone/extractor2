package html

import (
	"regexp"
	"strconv"
	"strings"
	"zhongguo/extractor2/context"

	"github.com/PuerkitoBio/goquery"
	"github.com/xlvector/dlog"
)

const (
	ARRAY_DEFINE = "_array"
	FIRST_DEFINE = "_first"
)

type HtmlSelector struct {
	Xpath string
	Attr  string
	Regex string
}

func NewHtmlSelector(v string) *HtmlSelector {
	ret := &HtmlSelector{}
	tks := strings.Split(v, ";")
	ret.Xpath = tks[0]
	if len(tks) > 1 {
		ret.Attr = tks[1]
	}
	if len(tks) > 2 {
		ret.Regex = tks[2]
	}
	return ret
}

func isEmptyParse(parseConfig map[string]interface{}) bool {
	for k, _ := range parseConfig {
		if !strings.HasPrefix(k, "_") {
			return false
		}
	}
	return true
}

func DoHtmlExtractor(parseConfig map[string]interface{}, val string, context *context.Context) interface{} {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(val))
	if err != nil {
		dlog.Warn("%s", err.Error())
		return nil
	}
	return doHtmlExtractorSelection(parseConfig, doc.First(), context)
}

func doHtmlExtractorSelection(parseConfig map[string]interface{}, selection *goquery.Selection, context *context.Context) interface{} {
	array_define, array_define_ok := parseConfig[ARRAY_DEFINE]
	first_define, first_define_ok := parseConfig[FIRST_DEFINE]
	rootSelection := selection
	if first_define_ok {
		rootSelection = selection.Find(first_define.(string))
	}
	if array_define_ok {
		rootSelection = rootSelection.Find(array_define.(string))
		isConfigEmpty := isEmptyParse(parseConfig)
		arrayRet := []interface{}{}
		rootSelection.Each(func(i int, stmp *goquery.Selection) {
			if isConfigEmpty {
				sub := stmp.Text()
				arrayRet = append(arrayRet, sub)
			} else {
				item := map[string]interface{}{}
				for key, parseVal := range parseConfig {
					if strings.HasPrefix(key, "_") {
						continue
					}
					if parseValQuery, ok := parseVal.(string); ok {
						item[key] = queryValue(parseValQuery, stmp, context)
					}
					if parseMapQuery, ok := parseVal.(map[string]interface{}); ok {
						item[key] = doHtmlExtractorSelection(parseMapQuery, stmp, context)
					}
				}
				arrayRet = append(arrayRet, item)
			}
		})
		return arrayRet
	}
	mapRet := map[string]interface{}{}
	for key, parseVal := range parseConfig {
		if strings.HasPrefix(key, "_") {
			continue
		}
		if parseValQuery, ok := parseVal.(string); ok {
			mapRet[key] = queryValue(parseValQuery, rootSelection, context)
		}
		if parseMapQuery, ok := parseVal.(map[string]interface{}); ok {
			mapRet[key] = doHtmlExtractorSelection(parseMapQuery, rootSelection, context)
		}
	}
	return mapRet
}

func doXpath(xpath string, s *goquery.Selection) *goquery.Selection {
	defer func() {
		if err := recover(); err != nil {
			dlog.Warn("doXpath Error:%v", err)
		}
	}()
	var b *goquery.Selection
	if strings.Contains(xpath, "[") && strings.Contains(xpath, "]") {
		xpathArr := []string{}
		for {
			start := strings.Index(xpath, "[")
			end := strings.Index(xpath, "]") + 1
			if start != -1 {
				if start > 0 {
					xpathArr = append(xpathArr, xpath[0:start])
				}
				xpathArr = append(xpathArr, xpath[start:end])
				xpath = xpath[end:]
			} else {
				xpathArr = append(xpathArr, xpath)
				break
			}
		}
		b = s
		for _, xpathItem := range xpathArr {
			if strings.HasPrefix(xpathItem, "[") && strings.HasSuffix(xpathItem, "]") {
				index, _ := strconv.Atoi(xpathItem[1 : len(xpathItem)-1])
				if index == -1 {
					b = b.Last()
				} else {
					b = b.Eq(index)
				}
			} else {
				b = b.Find(xpathItem)
			}
		}
	} else {
		b = s.Find(xpath)
	}
	return b
}

func queryValue(query string, s *goquery.Selection, context *context.Context) interface{} {
	defer func() {
		if err := recover(); err != nil {
			dlog.Warn("queryXpath Error:%v", err)
		}
	}()
	if strings.HasPrefix(query, "{{") && strings.HasSuffix(query, "}}") {
		return context.ParseTmplate(query)
	}
	if strings.Contains(query, "{{") && strings.Contains(query, "}}") {
		query = context.ParseTmplate(query)
	}
	var b *goquery.Selection
	b = s
	var text string
	//var err error
	htmlSelector := NewHtmlSelector(query)
	if len(htmlSelector.Xpath) > 0 {
		b = doXpath(query, s)
		text = b.Text()
	}
	if len(htmlSelector.Attr) > 0 {
		if htmlSelector.Attr == "html" {
			text, _ = b.Html()
		} else {
			text, _ = b.First().Attr(htmlSelector.Attr)
			if (htmlSelector.Attr == "href" || htmlSelector.Attr == "src") && strings.HasPrefix(text, "//") {
				text = "https:" + text
			}
			text = strings.TrimSpace(text)
		}
	}

	if len(htmlSelector.Regex) > 0 {
		reg := regexp.MustCompile(htmlSelector.Regex)
		result := reg.FindAllStringSubmatch(text, 1)
		if len(result) > 0 {
			group := result[0]
			if len(group) > 1 {
				text = group[1]
			} else {
				text = group[0]
			}
		} else {
			dlog.Warn("正则%s没有匹配到结果", htmlSelector.Regex)
			return nil
		}
	}
	return text
}
