package goplus

import (
	"hbuf/plugin/core/tpl"
	"hbuf/plugin/core/wire"
	"strings"
)

// MqContext mq 输出模板的 data 类型上下文。
type MqContext struct {
	Name    string // 驼峰名
	Stream  string // 下划线名
	Subject string // 下划线名转点分
}

// buildMq 为带 mq 注解的 data 类型构建 mq 模板上下文。
// 与进程内 printMqCode 一致:无 mq 注解时返回 nil。
func (c *Ctx) buildMq(imp Imports, data *wire.DataType) *MqContext {
	if _, ok := getTag(data.Tags, "mq"); !ok {
		return nil
	}
	imp.Import("context", "")
	imp.Import("github.com/nats-io/nats.go", "")
	imp.Import("github.com/nats-io/nats.go/jetstream", "")
	imp.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hmq", "")

	stream := tpl.UnderlineName(data.Name)
	return &MqContext{
		Name:    tpl.HumpName(data.Name),
		Stream:  stream,
		Subject: strings.ReplaceAll(stream, "_", "."),
	}
}
