package golang

import (
	"regexp"
	"strings"
	"testing"
)

type sqlParam struct {
	text string
}

func TestName(t *testing.T) {
	var allRex = regexp.MustCompile(`(\?{\w+})|(\${\w+})|\$|\?`)
	var text = "adfasfa?afasd$dfasd${fdasd}sfasdf?{afadfasd}"
	match := allRex.FindAllStringSubmatchIndex(text, -1)
	if nil != match {
		var index = 0
		buf := strings.Builder{}
		for _, item := range match {
			buf.WriteString(text[index:item[0]])
			index = item[0] + 1

		}
		if index < len(text) {
			buf.WriteString(text[index:])
		}
		text = buf.String()
	}
}
