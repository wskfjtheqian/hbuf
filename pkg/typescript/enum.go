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
	dst.Code("\tpublic readonly value: number\n\n")
	dst.Code("\tpublic readonly name: string\n\n")

	dst.Code("\tprivate constructor(value: number, name: string) {\n")
	dst.Code("\t\tthis.value = value;\n")
	dst.Code("\t\tthis.name = name;\n")
	dst.Code("\t}\n")

	dst.Code("\tpublic static valueOf(value: number): " + enumName + " {\n")
	dst.Code("\t	for (const i in " + enumName + ".values) {\n")
	dst.Code("\t		if (" + enumName + ".values[i].value == value) {\n")
	dst.Code("\t			return " + enumName + ".values[i];\n")
	dst.Code("\t		}\n")
	dst.Code("\t	}\n")
	dst.Code("\t	throw 'Get " + enumName + " by value error, value=${value}';\n")
	dst.Code("\t}\n\n")

	dst.Code("\tpublic static nameOf(name: string): " + enumName + " {\n")
	dst.Code("\t	for (const i in " + enumName + ".values) {\n")
	dst.Code("\t		if (" + enumName + ".values[i].name == name) {\n")
	dst.Code("\t			return " + enumName + ".values[i];\n")
	dst.Code("\t		}\n")
	dst.Code("\t	}\n")
	dst.Code("\t	throw 'Get " + enumName + " by name error, name=${name}';\n")
	dst.Code("\t}\n\n")

	for _, item := range typ.Items {
		if nil != item.Doc && 0 < len(item.Doc.Text()) {
			dst.Code("\t///" + item.Doc.Text())
		}
		itemName := build.StringToAllUpper(item.Name.Name)
		dst.Code("\tpublic static readonly " + itemName + " = new " + enumName + "(")
		dst.Code(item.Id.Value + ", \"" + build.StringToHumpName(item.Name.Name) + "\"")
		dst.Code(");\n\n")
	}
	dst.Code("\n")
	dst.Code("\tpublic static readonly values: " + enumName + "[] = [\n")
	for _, item := range typ.Items {
		dst.Code("\t\t" + enumName + "." + build.StringToAllUpper(item.Name.Name) + ",\n")
	}
	dst.Code("\t];\n\n")

	dst.Code("\ttoString(): string {\n")
	dst.Code("\t\treturn \"").Code(build.StringToFirstLower(enumName)).Code("Lang.\" + this.name\n")
	dst.Code("\t}\n\n")

	dst.Code("}\n")
}
