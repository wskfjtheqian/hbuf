package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func printEnumCode(dst *Writer, typ *ast.EnumType) {
	printEnum(dst, typ)
}

func printEnum(dst *Writer, typ *ast.EnumType) {
	enumName := build.StringToHumpName(typ.Name.Name)
	dst.Code("class " + enumName)
	dst.Code("{\n")
	dst.Code("  final int value;\n")
	dst.Code("  final String name;\n\n")

	dst.Code("  const " + enumName + "._(this.value, this.name);\n\n")

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
		dst.Code("  static const " + itemName + " = " + enumName + "._(" + item.Id.Value + ", '" + itemName + "');\n")
	}
	dst.Code("\n")
	dst.Code("  static const List<" + enumName + "> values = [\n")
	for _, item := range typ.Items {
		dst.Code("    " + build.StringToAllUpper(item.Name.Name) + ",\n")
	}
	dst.Code("  ];\n\n")

	dst.Code("}\n\n")
}
