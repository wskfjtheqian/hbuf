package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
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
	dst.Tab(0).Code("func (g ").Code(name).Code(") Publish(ctx context.Context) error {\n")
	dst.Tab(1).Code("return hmq.Publish(ctx, \"").Code(name).Code("\", &g)\n")
	dst.Tab(0).Code("}\n\n")
	return nil
}

func (b *Builder) printSubscribeCode(dst *build.Writer, data *ast.DataType, tag *ast.Tag) error {
	name := build.StringToHumpName(data.Name.Name)
	dst.Tab(0).Code("func (g ").Code(name).Code(") Subscribe(ctx context.Context, handler func(msg *").Code(name).Code(") error) (*nats.Subscription, error) {\n")
	dst.Tab(1).Code("return hmq.Subscribe(ctx, \"").Code(name).Code("\", handler)\n")
	dst.Tab(0).Code("}\n\n")
	return nil
}

func (b *Builder) printJetStreamPublishCode(dst *build.Writer, data *ast.DataType, tag *ast.Tag) error {
	name := build.StringToHumpName(data.Name.Name)
	dst.Tab(0).Code("func (g ").Code(name).Code(") JsPublish(ctx context.Context, stream string) (*jetstream.PubAck, error) {\n")
	dst.Tab(1).Code("return hmq.JetStreamPublish(ctx, stream, \"").Code(name).Code("\", &g)\n")
	dst.Tab(0).Code("}\n\n")
	return nil
}

func (b *Builder) printJetStreamSubscribeCode(dst *build.Writer, data *ast.DataType, tag *ast.Tag) error {
	name := build.StringToHumpName(data.Name.Name)
	dst.Tab(0).Code("func (g ").Code(name).Code(") JsSubscribe(ctx context.Context, stream string, durable string, handler func(msg *").Code(name).Code(") error) error {\n")
	dst.Tab(1).Code("return hmq.JetStreamSubscribe(ctx, stream, \"").Code(name).Code("\", durable, handler)\n")
	dst.Tab(0).Code("}\n\n")
	return nil
}
