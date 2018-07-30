package context

type Context struct {
	DoTemplateFunc func(template string) string //去执行业务真正的context
}

func NewContext(doTemplateFunc func(template string) string) *Context {
	instance := Context{}
	instance.DoTemplateFunc = doTemplateFunc
	return &instance
}

func (self *Context) ParseTmplate(tmplate string) string {
	return self.DoTemplateFunc(tmplate)
}
