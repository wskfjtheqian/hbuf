package java

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printEnumCode(dst *build.Writer, typ *ast.EnumType) {
	b.printEnum(dst, typ)
}

func (b *Builder) printEnum(dst *build.Writer, typ *ast.EnumType) {
	enumName := build.StringToHumpName(typ.Name.Name)
	_, isUi := build.GetTag(typ.Tags, "ui")
	if isUi {
		b.getPackage(dst, typ.Name, "ui")
	}
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("///" + typ.Doc.Text())
	}
	dst.Code("\tclass " + enumName)
	dst.Code("\t{\n")
	dst.Code("\t\tfinal int value;\n")
	dst.Code("\t\tfinal String name;\n\n")

	dst.Code("\t\tprivate " + enumName + "(int value, String name){\n")
	dst.Code("\t\t\tthis.value = value;\n")
	dst.Code("\t\t\tthis.name = name;\n")
	dst.Code("\t\t}\n\n")

	dst.Code("\t\t@Override\n")
	dst.Code("\t\tpublic boolean equals(Object o) {\n")
	dst.Code("\t\t\tif (this == o) return true;\n")
	dst.Code("\t\t\tif (o == null || getClass() != o.getClass()) return false;\n")
	dst.Code("\t\t\tTest test = (Test) o;\n")
	dst.Code("\t\t\treturn value == test.value;\n")
	dst.Code("\t\t}\n\n")

	dst.Code("\t\t@Override\n")
	dst.Code("\t\tpublic int hashCode() {\n")
	dst.Code("\t\t\treturn this.value;\n")
	dst.Code("\t\t}\n\n")

	dst.Code("\t\tpublic static " + enumName + " valueOf(int value) {\n")
	dst.Code("\t\t\tfor (" + enumName + " item : values) {\n")
	dst.Code("\t\t\t\tif (item.value == value) {\n")
	dst.Code("\t\t\t\t\treturn item;\n")
	dst.Code("\t\t\t\t}\n")
	dst.Code("\t\t\t}\n")
	dst.Code("\t\t\tthrow new RuntimeException(\"Get Test by value error, value=\" + value);\n")
	dst.Code("\t\t}\n\n")

	dst.Code("\t\tpublic static " + enumName + " nameOf(String name) {\n")
	dst.Code("\t\t\tfor (" + enumName + " item : values) {\n")
	dst.Code("\t\t\t\tif (item.name == name) {\n")
	dst.Code("\t\t\t\t\treturn item;\n")
	dst.Code("\t\t\t\t}\n")
	dst.Code("\t\t\t}\n")
	dst.Code("\t\t\tthrow new RuntimeException(\"Get Test by name error, name=\" + name);\n")
	dst.Code("\t\t}\n\n")

	for _, item := range typ.Items {
		if nil != item.Doc && 0 < len(item.Doc.Text()) {
			dst.Code("///" + item.Doc.Text())
		}
		itemName := build.StringToAllUpper(item.Name.Name)
		dst.Code("\t\tstatic final " + enumName + " " + itemName + " = new " + enumName + "(" + item.Id.Value + ", \"" + itemName + "\"")
		if isUi {
			dst.Code(", (context) => " + enumName + "Localizations.of(context)." + itemName)
		}
		dst.Code(");\n")
	}
	dst.Code("\n")
	dst.Code("\t\tstatic final " + enumName + "[] values = new " + enumName + "[]{\n")
	for _, item := range typ.Items {
		dst.Code("\t\t\t" + build.StringToAllUpper(item.Name.Name) + ",\n")
	}
	dst.Code("\t\t};\n\n")

	dst.Code("\t\t@Override\n")
	dst.Code("\t\tpublic String toString() {\n")
	dst.Code("\t\t\treturn name;\n")
	dst.Code("\t\t}\n\n")

	dst.Code("\t}\n\n")

}
