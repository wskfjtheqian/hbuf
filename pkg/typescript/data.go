package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printDataCode(dst *build.Writer, typ *ast.DataType) error {
	return b.printData(dst, typ)
}

func (b *Builder) printData(dst *build.Writer, typ *ast.DataType) error {
	// desc 单例（NewSyncDescriptor 包裹，延迟构建，支持循环引用）
	err := b.printDataDescriptor(dst, typ)
	if err != nil {
		return err
	}

	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("///" + typ.Doc.Text())
	}
	uName := className(typ.Name.Name)
	dst.Import("hbuf_ts", "type HData", "type Descriptor")
	dst.Code("export class " + uName + " implements HData")
	if nil != typ.Extends {
		b.printExtend(dst, typ.Extends, true)
	}
	dst.Code(" {\n")

	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Tab(1).Code("///" + field.Doc.Text())
		}
		dst.Tab(1).Code("")
		dst.Code(build.StringToFirstLower(field.Name.Name) + ": ")
		b.printType(dst, field.Type, false, false, false)
		dst.Code(" = ")
		b.printDefault(dst, field.Type, false)
		dst.Code(";\n\n")
		return nil
	})
	if err != nil {
		return err
	}

	dName := descName(typ.Name.Name)

	dst.Tab(1).Code("descriptors(): Descriptor {\n")
	dst.Tab(2).Code("return " + dName + "\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public static fromMap(_json: Record<string, any>): " + uName + "{\n")
	dst.Tab(2).Code("return ").Code(dName).Code(".fromJson(_json) as " + uName + "\n")
	dst.Tab(1).Code("}\n\n")

	dst.Code("\n")
	dst.Tab(1).Code("public toMap(): Record<string, any> {\n")
	dst.Tab(2).Code("return ").Code(dName).Code(".toJson(this) as Record<string, any> \n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public clone(): ").Code(uName).Code(" {\n")
	dst.Tab(2).Code("return ").Code(dName).Code(".clone(this) as ").Code(uName).Code(" \n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public $getChangeField(val: ").Code(uName).Code(", ignore: string[] = []): ")
	dst.Code("{ change : ").Code(uName).Code(", ").Code("fields: string[]").Code(" } {\n")
	dst.Tab(2).Code("return ").Code(dName).Code(".getChangeField(this, val, ignore)\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public static $convertField(fields: string[]): string[] {\n")
	dst.Tab(2).Code("return fields.map((v)=>").Code(dName).Code(".convertField(v))\n")
	dst.Tab(1).Code("}\n\n")
	dst.Code("}\n\n")

	return nil
}

func (b *Builder) printExtend(dst *build.Writer, extends []*ast.Extends, start bool) {
	for i, v := range extends {
		if 0 != i || start {
			dst.Code(", ")
		}

		b.getPackage(dst, v.Name, "", true, false, "", "", nil)
		dst.Code(build.StringToHumpName(v.Name.Name))
	}
}
