package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printEnumCode(dst *build.Writer, typ *ast.EnumType) {
	dst.Import("package:flutter/widgets.dart", "")
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
	dst.Code("class " + enumName)
	dst.Code("{\n")
	dst.Tab(1).Code("final int value;\n\n")
	dst.Tab(1).Code("final String name;\n\n")
	dst.Tab(1).Code("final String Function(BuildContext context)? _onText;\n\n")
	dst.Tab(1).Code("const " + enumName + "._(this.value, this.name, [this._onText]);\n\n")

	dst.Tab(1).Code("@override\n")
	dst.Tab(1).Code("bool operator ==(Object other) =>\n")
	dst.Tab(3).Code("identical(this, other) ||\n")
	dst.Tab(3).Code("other is " + enumName + " &&\n")
	dst.Tab(5).Code("runtimeType == other.runtimeType &&\n")
	dst.Tab(5).Code("value == other.value;\n\n")

	dst.Tab(1).Code("@override\n")
	dst.Tab(1).Code("int get hashCode => value.hashCode;\n\n")

	dst.Tab(1).Code("static " + enumName + " valueOf(int value) {\n")
	dst.Tab(1).Code("	for (var item in values) {\n")
	dst.Tab(1).Code("		if (item.value == value) {\n")
	dst.Tab(1).Code("			return item;\n")
	dst.Tab(1).Code("		}\n")
	dst.Tab(1).Code("	}\n")
	dst.Tab(1).Code("	throw 'Get " + enumName + " by value error, value=$value';\n")
	dst.Tab(1).Code("}\n\n")

	dst.Tab(1).Code("static " + enumName + " nameOf(String name) {\n")
	dst.Tab(1).Code("	for (var item in values) {\n")
	dst.Tab(1).Code("		if (item.name == name) {\n")
	dst.Tab(1).Code("			return item;\n")
	dst.Tab(1).Code("		}\n")
	dst.Tab(1).Code("	}\n")
	dst.Tab(1).Code("	throw 'Get " + enumName + " by name error, name=$name';\n")
	dst.Tab(1).Code("}\n\n")

	for _, item := range typ.Items {
		if nil != item.Doc && 0 < len(item.Doc.Text()) {
			dst.Tab(1).Code("///" + item.Doc.Text())
		}
		itemName := build.StringToAllUpper(item.Name.Name)
		dst.Tab(1).Code("static final " + itemName + " = " + enumName + "._(" + item.Id.Value + ", '" + build.StringToHumpName(item.Name.Name) + "'")
		if isUi {
			dst.Code(", (context) => " + enumName + "Localizations.of(context)." + itemName)
		}
		dst.Code(");\n\n")
	}
	dst.Code("\n")
	dst.Tab(1).Code("static final List<" + enumName + "> values = [\n")
	for _, item := range typ.Items {
		dst.Tab(2).Code("" + build.StringToAllUpper(item.Name.Name) + ",\n")
	}
	dst.Tab(1).Code("];\n\n")

	dst.Tab(1).Code("@override\n")
	dst.Tab(1).Code("String toString() {\n")
	dst.Tab(2).Code("return name;\n")
	dst.Tab(1).Code("}\n")

	dst.Tab(1).Code("String toText(BuildContext context) {\n")
	dst.Tab(2).Code("return _onText?.call(context) ?? name;\n")
	dst.Tab(1).Code("}\n")
	dst.Code("}\n\n")

}
