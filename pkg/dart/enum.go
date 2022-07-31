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

	dst.Code("class " + enumName)
	dst.Code("{\n")
	dst.Code("  final int value;\n")
	dst.Code("  final String name;\n\n")
	dst.Code("  final String Function(BuildContext context)? _onText;\n")
	dst.Code("  const " + enumName + "._(this.value, this.name, [this._onText]);\n\n")

	dst.Code("  @override\n")
	dst.Code("  bool operator ==(Object other) =>\n")
	dst.Code("      identical(this, other) ||\n")
	dst.Code("      other is " + enumName + " &&\n")
	dst.Code("          runtimeType == other.runtimeType &&\n")
	dst.Code("          value == other.value;\n\n")

	dst.Code("  @override\n")
	dst.Code("  int get hashCode => value.hashCode;\n\n")

	dst.Code("  static " + enumName + " valueOf(int value) {\n")
	dst.Code("  	for (var item in values) {\n")
	dst.Code("  		if (item.value == value) {\n")
	dst.Code("  			return item;\n")
	dst.Code("  		}\n")
	dst.Code("  	}\n")
	dst.Code("  	throw 'Get " + enumName + " by value error, value=$value';\n")
	dst.Code("  }\n\n")

	dst.Code("  static " + enumName + " nameOf(String name) {\n")
	dst.Code("  	for (var item in values) {\n")
	dst.Code("  		if (item.name == name) {\n")
	dst.Code("  			return item;\n")
	dst.Code("  		}\n")
	dst.Code("  	}\n")
	dst.Code("  	throw 'Get " + enumName + " by name error, name=$name';\n")
	dst.Code("  }\n\n")

	for _, item := range typ.Items {
		itemName := build.StringToAllUpper(item.Name.Name)
		dst.Code("  static final " + itemName + " = " + enumName + "._(" + item.Id.Value + ", '" + itemName + "'")
		if isUi {
			dst.Code(", (context) => " + enumName + "Localizations.of(context)." + itemName)
		}
		dst.Code(");\n")
	}
	dst.Code("\n")
	dst.Code("  static final List<" + enumName + "> values = [\n")
	for _, item := range typ.Items {
		dst.Code("    " + build.StringToAllUpper(item.Name.Name) + ",\n")
	}
	dst.Code("  ];\n\n")

	dst.Code("  @override\n")
	dst.Code("  String toString() {\n")
	dst.Code("    return name;\n")
	dst.Code("  }\n")

	dst.Code("  String toText(BuildContext context) {\n")
	dst.Code("    return _onText?.call(context) ?? name;\n")
	dst.Code("  }\n")
	dst.Code("}\n\n")

}
