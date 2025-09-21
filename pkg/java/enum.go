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
	dst.Tab(1).Code("class " + enumName)
	dst.Tab(1).Code("{\n")
	dst.Tab(2).Code("final int value;\n")
	dst.Tab(2).Code("final String name;\n\n")

	dst.Tab(2).Code("private " + enumName + "(int value, String name){\n")
	dst.Tab(3).Code("this.value = value;\n")
	dst.Tab(3).Code("this.name = name;\n")
	dst.Tab(2).Code("}\n\n")

	dst.Tab(2).Code("@Override\n")
	dst.Tab(2).Code("public boolean equals(Object o) {\n")
	dst.Tab(3).Code("if (this == o) return true;\n")
	dst.Tab(3).Code("if (o == null || getClass() != o.getClass()) return false;\n")
	dst.Tab(3).Code("Test test = (Test) o;\n")
	dst.Tab(3).Code("return value == test.value;\n")
	dst.Tab(2).Code("}\n\n")

	dst.Tab(2).Code("@Override\n")
	dst.Tab(2).Code("public int hashCode() {\n")
	dst.Tab(3).Code("return this.value;\n")
	dst.Tab(2).Code("}\n\n")

	dst.Tab(2).Code("public static " + enumName + " valueOf(int value) {\n")
	dst.Tab(3).Code("for (" + enumName + " item : values) {\n")
	dst.Tab(4).Code("if (item.value == value) {\n")
	dst.Tab(5).Code("return item;\n")
	dst.Tab(4).Code("}\n")
	dst.Tab(3).Code("}\n")
	dst.Tab(3).Code("throw new RuntimeException(\"Get Test by value error, value=\" + value);\n")
	dst.Tab(2).Code("}\n\n")

	dst.Tab(2).Code("public static " + enumName + " nameOf(String name) {\n")
	dst.Tab(3).Code("for (" + enumName + " item : values) {\n")
	dst.Tab(4).Code("if (item.name == name) {\n")
	dst.Tab(5).Code("return item;\n")
	dst.Tab(4).Code("}\n")
	dst.Tab(3).Code("}\n")
	dst.Tab(3).Code("throw new RuntimeException(\"Get Test by name error, name=\" + name);\n")
	dst.Tab(2).Code("}\n\n")

	for _, item := range typ.Items {
		if nil != item.Doc && 0 < len(item.Doc.Text()) {
			dst.Code("///" + item.Doc.Text())
		}
		itemName := build.StringToAllUpper(item.Name.Name)
		dst.Tab(2).Code("static final " + enumName + " " + itemName + " = new " + enumName + "(" + item.Id.Value + ", \"" + itemName + "\"")
		if isUi {
			dst.Code(", (context) => " + enumName + "Localizations.of(context)." + itemName)
		}
		dst.Code(");\n")
	}
	dst.Code("\n")
	dst.Tab(2).Code("static final " + enumName + "[] values = new " + enumName + "[]{\n")
	for _, item := range typ.Items {
		dst.Tab(3).Code("" + build.StringToAllUpper(item.Name.Name) + ",\n")
	}
	dst.Tab(2).Code("};\n\n")

	dst.Tab(2).Code("@Override\n")
	dst.Tab(2).Code("public String toString() {\n")
	dst.Tab(3).Code("return name;\n")
	dst.Tab(2).Code("}\n\n")

	dst.Tab(1).Code("}\n\n")

}
