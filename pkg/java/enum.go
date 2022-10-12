package java

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
	dst.Code("\tfinal int value;\n")
	dst.Code("\tfinal String name;\n\n")
	dst.Code("\tfinal String Function(BuildContext context)? _onText;\n")
	dst.Code("\tconst " + enumName + "._(this.value, this.name, [this._onText]);\n\n")

	dst.Code("\t@override\n")
	dst.Code("\tbool operator ==(Object other) =>\n")
	dst.Code("\t\t\tidentical(this, other) ||\n")
	dst.Code("\t\t\tother is " + enumName + " &&\n")
	dst.Code("\t\t\t\t\truntimeType == other.runtimeType &&\n")
	dst.Code("\t\t\t\t\tvalue == other.value;\n\n")

	dst.Code("\t@override\n")
	dst.Code("\tint get hashCode => value.hashCode;\n\n")

	dst.Code("\tstatic " + enumName + " valueOf(int value) {\n")
	dst.Code("\t	for (var item in values) {\n")
	dst.Code("\t		if (item.value == value) {\n")
	dst.Code("\t			return item;\n")
	dst.Code("\t		}\n")
	dst.Code("\t	}\n")
	dst.Code("\t	throw 'Get " + enumName + " by value error, value=$value';\n")
	dst.Code("\t}\n\n")

	dst.Code("\tstatic " + enumName + " nameOf(String name) {\n")
	dst.Code("\t	for (var item in values) {\n")
	dst.Code("\t		if (item.name == name) {\n")
	dst.Code("\t			return item;\n")
	dst.Code("\t		}\n")
	dst.Code("\t	}\n")
	dst.Code("\t	throw 'Get " + enumName + " by name error, name=$name';\n")
	dst.Code("\t}\n\n")

	for _, item := range typ.Items {
		if nil != item.Doc && 0 < len(item.Doc.Text()) {
			dst.Code("///" + item.Doc.Text())
		}
		itemName := build.StringToAllUpper(item.Name.Name)
		dst.Code("\tstatic final " + itemName + " = " + enumName + "._(" + item.Id.Value + ", '" + itemName + "'")
		if isUi {
			dst.Code(", (context) => " + enumName + "Localizations.of(context)." + itemName)
		}
		dst.Code(");\n")
	}
	dst.Code("\n")
	dst.Code("\tstatic final List<" + enumName + "> values = [\n")
	for _, item := range typ.Items {
		dst.Code("\t\t" + build.StringToAllUpper(item.Name.Name) + ",\n")
	}
	dst.Code("\t];\n\n")

	dst.Code("\t@override\n")
	dst.Code("\tString toString() {\n")
	dst.Code("\t\treturn name;\n")
	dst.Code("\t}\n")

	dst.Code("\tString toText(BuildContext context) {\n")
	dst.Code("\t\treturn _onText?.call(context) ?? name;\n")
	dst.Code("\t}\n")
	dst.Code("}\n\n")

}
