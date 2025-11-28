package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strings"
)

func (b *Builder) printMqCode(dst *build.Writer, data *ast.DataType) error {
	tag, ok := build.GetTag(data.Tags, "mq")
	if !ok {
		return nil
	}
	dst.Import("context", "")
	dst.Import("github.com/nats-io/nats.go", "")
	dst.Import("github.com/nats-io/nats.go/jetstream", "")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hmq", "")

	err := b.printPublishCode(dst, data, tag)
	if err != nil {
		return err
	}
	err = b.printSubscribeCode(dst, data, tag)
	if err != nil {
		return err
	}
	err = b.printJetStreamPublishCode(dst, data, tag)
	if err != nil {
		return err
	}
	err = b.printJetStreamSubscribeCode(dst, data, tag)
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) printPublishCode(dst *build.Writer, data *ast.DataType, tag *ast.Tag) error {
	name := build.StringToHumpName(data.Name.Name)
	subject := strings.ReplaceAll(build.StringToUnderlineName(data.Name.Name), "_", ".")
	dst.Tab(0).Code("func (g ").Code(name).Code(") Publish(ctx context.Context, subject string) error {\n")
	dst.Tab(1).Code("return hmq.Publish(ctx, \"").Code(subject).Code(".\"+subject, &g)\n")
	dst.Tab(0).Code("}\n\n")
	return nil
}

func (b *Builder) printSubscribeCode(dst *build.Writer, data *ast.DataType, tag *ast.Tag) error {
	name := build.StringToHumpName(data.Name.Name)
	subject := strings.ReplaceAll(build.StringToUnderlineName(data.Name.Name), "_", ".")
	dst.Tab(0).Code("func (g ").Code(name).Code(") Subscribe(ctx context.Context, subject string, handler func(ctx context.Context, subject string, msg *").Code(name).Code(") error) (*nats.Subscription, error) {\n")
	dst.Tab(1).Code("return hmq.Subscribe(ctx, \"").Code(subject).Code(".\"+subject, handler)\n")
	dst.Tab(0).Code("}\n\n")
	return nil
}

func (b *Builder) printJetStreamPublishCode(dst *build.Writer, data *ast.DataType, tag *ast.Tag) error {
	name := build.StringToHumpName(data.Name.Name)
	stream := build.StringToUnderlineName(data.Name.Name)
	subject := strings.ReplaceAll(build.StringToUnderlineName(data.Name.Name), "_", ".")
	dst.Tab(0).Code("func (g ").Code(name).Code(") JsPublish(ctx context.Context, subject string, options ...hmq.PublishOption) (*jetstream.PubAck, error) {\n")
	dst.Tab(1).Code("return hmq.JetStreamPublish(ctx, \"stream_").Code(stream).Code("\", \"").Code(subject).Code(".\"+subject, &g, options...)\n")
	dst.Tab(0).Code("}\n\n")
	return nil
}

func (b *Builder) printJetStreamSubscribeCode(dst *build.Writer, data *ast.DataType, tag *ast.Tag) error {
	name := build.StringToHumpName(data.Name.Name)
	stream := build.StringToUnderlineName(data.Name.Name)
	subject := strings.ReplaceAll(build.StringToUnderlineName(data.Name.Name), "_", ".")
	dst.Tab(0).Code("func (g ").Code(name).Code(") JsSubscribe(ctx context.Context, subject string, durable string, handler func(ctx context.Context, subject, msgId string, msg *").Code(name).Code(") error, options ...hmq.SubscribeOption) error {\n")
	dst.Tab(1).Code("return hmq.JetStreamSubscribe(ctx, \"stream_").Code(stream).Code("\", \"").Code(subject).Code(".\"+subject, durable, handler, options...)\n")
	dst.Tab(0).Code("}\n\n")
	return nil
}
