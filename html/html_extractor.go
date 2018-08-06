package html

import (
	"strconv"
	"strings"

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
}

func NewHtmlSelector(v string) *HtmlSelector {
	ret := &HtmlSelector{}
	tks := strings.Split(v, ";")
	ret.Xpath = tks[0]
	if len(tks) > 1 {
		ret.Attr = tks[1]
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

//解析多个结果
func DoHtmlExtractor(array_define, val string) []interface{} {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(val))
	if err != nil {
		dlog.Warn("%s", err.Error())
		return nil
	}
	var ret []interface{}
	rootSelection := doc.Find(array_define)
	rootSelection.Each(func(i int, stmp *goquery.Selection) {
		htxt, _ := stmp.Html()
		ret = append(ret, htxt)
	})
	return ret
}

//解析单个结果
func DoHtmlOneExtractor(parseConfig string, val string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(val))
	if err != nil {
		dlog.Warn("%s", err.Error())
		return ""
	}
	return queryValue(parseConfig, doc.First())
}

func doHtmlExtractorSelection(parseConfig map[string]interface{}, selection *goquery.Selection) interface{} {
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
						item[key] = queryValue(parseValQuery, stmp)
					}
					if parseMapQuery, ok := parseVal.(map[string]interface{}); ok {
						item[key] = doHtmlExtractorSelection(parseMapQuery, stmp)
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
			mapRet[key] = queryValue(parseValQuery, rootSelection)
		}
		if parseMapQuery, ok := parseVal.(map[string]interface{}); ok {
			mapRet[key] = doHtmlExtractorSelection(parseMapQuery, rootSelection)
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
				if strings.TrimSpace(xpath) != "" {
					xpathArr = append(xpathArr, xpath)
				}
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

func queryValue(query string, s *goquery.Selection) string {
	defer func() {
		if err := recover(); err != nil {
			dlog.Warn("queryXpath Error:%v", err)
		}
	}()
	var b *goquery.Selection
	b = s
	var text string
	//var err error
	htmlSelector := NewHtmlSelector(query)
	if len(htmlSelector.Xpath) > 0 {
		b = doXpath(htmlSelector.Xpath, s)
		text, _ = b.Html()
	}
	if len(htmlSelector.Attr) > 0 {
		if htmlSelector.Attr == "html" {
			text, _ = b.Html()
		} else {
			text, _ = b.Attr(htmlSelector.Attr)
			if (htmlSelector.Attr == "href" || htmlSelector.Attr == "src") && strings.HasPrefix(text, "//") {
				text = "https:" + text
			}
			text = strings.TrimSpace(text)
		}
	}
	return text
}
