package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printEnumCode(dst *build.Writer, typ *ast.EnumType) {
	b.printEnum(dst, typ)
}

func (b *Builder) printEnum(dst *build.Writer, typ *ast.EnumType) {
	enumName := build.StringToHumpName(typ.Name.Name)
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("///" + typ.Doc.Text())
	}
	dst.Code("export class " + enumName)
	dst.Code("{\n")
	dst.Tab(1).Code("public readonly value: number\n\n")
	dst.Tab(1).Code("public readonly name: string\n\n")

	dst.Tab(1).Code("private constructor(value: number, name: string) {\n")
	dst.Tab(2).Code("this.value = value;\n")
	dst.Tab(2).Code("this.name = name;\n")
	dst.Tab(1).Code("}\n")

	dst.Tab(1).Code("public static valueOf(value: number): " + enumName + " {\n")
	dst.Tab(1).Code("	for (const i in " + enumName + ".values) {\n")
	dst.Tab(1).Code("		if (" + enumName + ".values[i].value == value) {\n")
	dst.Tab(1).Code("			return " + enumName + ".values[i];\n")
	dst.Tab(1).Code("		}\n")
	dst.Tab(1).Code("	}\n")
	dst.Tab(1).Code("	throw 'Get " + enumName + " by value error, value=${value}';\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public static nameOf(name: string): " + enumName + " {\n")
	dst.Tab(1).Code("	for (const i in " + enumName + ".values) {\n")
	dst.Tab(1).Code("		if (" + enumName + ".values[i].name == name) {\n")
	dst.Tab(1).Code("			return " + enumName + ".values[i];\n")
	dst.Tab(1).Code("		}\n")
	dst.Tab(1).Code("	}\n")
	dst.Tab(1).Code("	throw 'Get " + enumName + " by name error, name=${name}';\n")
	dst.Tab(1).Code("}\n\n")

	for _, item := range typ.Items {
		if nil != item.Doc && 0 < len(item.Doc.Text()) {
			dst.Tab(1).Code("///" + item.Doc.Text())
		}
		itemName := build.StringToAllUpper(item.Name.Name)
		dst.Tab(1).Code("public static readonly " + itemName + " = new " + enumName + "(")
		dst.Code(item.Id.Value + ", \"" + build.StringToHumpName(item.Name.Name) + "\"")
		dst.Code(");\n\n")
	}
	dst.Code("\n")
	dst.Tab(1).Code("public static readonly values: " + enumName + "[] = [\n")
	for _, item := range typ.Items {
		dst.Tab(2).Code("" + enumName + "." + build.StringToAllUpper(item.Name.Name) + ",\n")
	}
	dst.Tab(1).Code("];\n\n")

	dst.Tab(1).Code("toString(): string {\n")
	dst.Tab(2).Code("return \"").Code(build.StringToFirstLower(enumName)).Code("Lang.\" + this.name\n")
	dst.Tab(1).Code("}\n\n")

	dst.Code("}\n")
}
