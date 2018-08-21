package json

import (
	"strings"
)

func FilterJSONP(s string) string {
	p0 := strings.Index(s, "{")
	p00 := strings.Index(s, "[")
	if p0 > p00 {
		p0 = p00
	}
	p1 := strings.Index(s, "(")
	if p1 < 0 {
		return s
	} else if p1 >= 0 && p1 > p0 {
		return s
	}
	p1 += 1
	p2 := strings.LastIndex(s, ")")
	if p2 <= p1 {
		return s
	}
	return strings.Trim(s[p1:p2], "\"")
}
