package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strings"
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
	dst.Tab(1).Code("public readonly cssClass: string\n\n")

	dst.Tab(1).Code("private constructor(value: number, name: string, cssClass?: string) {\n")
	dst.Tab(2).Code("this.value = value\n")
	dst.Tab(2).Code("this.name = name\n")
	dst.Tab(2).Code("this.cssClass = cssClass || ''\n")
	dst.Tab(1).Code("}\n")

	dst.Tab(1).Code("public static valueOf(value: number): " + enumName + " {\n")
	dst.Tab(1).Code("	for (const v of " + enumName + ".values) {\n")
	dst.Tab(1).Code("		if (v.value == value) {\n")
	dst.Tab(1).Code("			return v\n")
	dst.Tab(1).Code("		}\n")
	dst.Tab(1).Code("	}\n")
	dst.Tab(1).Code("	return { value: value, name: `Unknown ${value}`, cssClass: `` }\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("public static nameOf(name: string): " + enumName + " {\n")
	dst.Tab(1).Code("	for (const v of " + enumName + ".values) {\n")
	dst.Tab(1).Code("		if (v.name == name) {\n")
	dst.Tab(1).Code("			return v\n")
	dst.Tab(1).Code("		}\n")
	dst.Tab(1).Code("	}\n")
	dst.Tab(1).Code("	return { value: -1, name: name, cssClass: ``  }\n")
	dst.Tab(1).Code("}\n\n")

	for _, item := range typ.Items {
		if nil != item.Doc && 0 < len(item.Doc.Text()) {
			dst.Tab(1).Code("///" + item.Doc.Text())
		}
		itemName := build.StringToAllUpper(item.Name.Name)
		dst.Tab(1).Code("public static readonly " + itemName + " = new " + enumName + "(")
		dst.Code(item.Id.Value + ", \"" + build.StringToHumpName(item.Name.Name) + "\"")

		tag, ok := build.GetTag(item.Tags, "ui")
		if ok && tag.KV != nil {
			classList := make([]string, 0)
			for _, item := range tag.KV {
				if item.Name.Name == "class" {
					for _, value := range item.Values {
						classList = append(classList, value.Value[1:len(value.Value)-1])
					}
				}
			}
			if 0 < len(classList) {
				dst.Code(", \"").Code(strings.Join(classList, " ")).Code("\"")
			}
		}
		dst.Code(")\n\n")
	}
	dst.Code("\n")
	dst.Tab(1).Code("public static readonly values: " + enumName + "[] = [\n")
	for _, item := range typ.Items {
		dst.Tab(2).Code("" + enumName + "." + build.StringToAllUpper(item.Name.Name) + ",\n")
	}
	dst.Tab(1).Code("]\n\n")

	dst.Tab(1).Code("toString(): string {\n")
	dst.Tab(2).Code("return \"").Code(build.StringToFirstLower(enumName)).Code("Lang.\" + this.name\n")
	dst.Tab(1).Code("}\n\n")

	dst.Code("}\n")
}
