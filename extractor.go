package extractor2

import (
	"zhongguo/extractor2/context"
	"zhongguo/extractor2/html"
	"zhongguo/extractor2/json"
	estring "zhongguo/extractor2/string"
)

type Extractor struct {
	context *context.Context
}

func NewExtractor(doTemplateFunc func(template string) string) *Extractor {
	instance := Extractor{}
	instance.context = context.NewContext(doTemplateFunc)
	return &instance
}

func (self *Extractor) Do(parseConfig map[string]interface{}, body string) interface{} {
	contentType := DetectContentType(body)
	if contentType == 0 {
		return html.DoHtmlExtractor(parseConfig, body, self.context)
	} else if contentType == 1 {
		return json.DoJsonExtractor(parseConfig, body, self.context)
	} else if contentType == 2 {
		return estring.DoStringExtractor(parseConfig, body, self.context)
	}
	return nil
}
