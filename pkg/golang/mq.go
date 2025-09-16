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
	dst.Import("encoding/json", "")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/erro", "")
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/mq", "")

	err := b.printPublishMsgCode(dst, data, tag)
	if err != nil {
		return err
	}
	err = b.printSubscribeCode(dst, data, tag)
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) printPublishMsgCode(dst *build.Writer, data *ast.DataType, tag *ast.Tag) error {
	name := build.StringToHumpName(data.Name.Name)
	dst.Tab(0).Code("func (g ").Code(name).Code(") PublishMsg(ctx context.Context) error {\n")
	dst.Tab(1).Code("bytes, err := json.Marshal(&g)\n")
	dst.Tab(1).Code("if err != nil {\n")
	dst.Tab(2).Code("return erro.Wrap(err)\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("err = mq.GET(ctx).PublishMsg(\"").Code(name).Code("\", bytes)\n")
	dst.Tab(1).Code("if err != nil {\n")
	dst.Tab(2).Code("return err\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("return nil\n")
	dst.Tab(0).Code("}\n\n")
	return nil
}

func (b *Builder) printSubscribeCode(dst *build.Writer, data *ast.DataType, tag *ast.Tag) error {
	name := build.StringToHumpName(data.Name.Name)
	dst.Tab(0).Code("func (g ").Code(name).Code(") Subscribe(ctx context.Context, handler func(msg *").Code(name).Code(") error) error {\n")
	dst.Tab(1).Code("err := mq.GET(ctx).Subscribe(\"").Code(name).Code("\", func(data []byte) error {\n")
	dst.Tab(2).Code("var req ").Code(name).Code("\n")
	dst.Tab(2).Code("err := json.Unmarshal(data, &req)\n")
	dst.Tab(2).Code("if err != nil {\n")
	dst.Tab(3).Code("return erro.Wrap(err)\n")
	dst.Tab(2).Code("}\n")
	dst.Tab(2).Code("return handler(&req)\n")
	dst.Tab(1).Code("})\n")
	dst.Tab(1).Code("if err != nil {\n")
	dst.Tab(2).Code("return err\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("return nil\n")
	dst.Tab(0).Code("}\n\n")
	return nil
}
