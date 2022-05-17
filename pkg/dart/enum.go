package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"io"
)

func printEnum(dst io.Writer, typ *ast.EnumType) {
	enumName := build.StringToHumpName(typ.Name.Name)
	_, _ = dst.Write([]byte("class " + enumName))
	_, _ = dst.Write([]byte("{\n"))
	_, _ = dst.Write([]byte("  final int value;\n"))
	_, _ = dst.Write([]byte("  final String name;\n\n"))

	_, _ = dst.Write([]byte("  const " + enumName + "._(this.value, this.name);\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  bool operator ==(Object other) =>\n"))
	_, _ = dst.Write([]byte("      identical(this, other) ||\n"))
	_, _ = dst.Write([]byte("      other is " + enumName + " &&\n"))
	_, _ = dst.Write([]byte("          runtimeType == other.runtimeType &&\n"))
	_, _ = dst.Write([]byte("          value == other.value;\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  int get hashCode => value.hashCode;\n\n"))

	_, _ = dst.Write([]byte("  static " + enumName + " valueOf(int value) {\n"))
	_, _ = dst.Write([]byte("  	for (var item in values) {\n"))
	_, _ = dst.Write([]byte("  		if (item.value == value) {\n"))
	_, _ = dst.Write([]byte("  			return item;\n"))
	_, _ = dst.Write([]byte("  		}\n"))
	_, _ = dst.Write([]byte("  	}\n"))
	_, _ = dst.Write([]byte("  	throw 'Get " + enumName + " by value error, value=$value';\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("  static " + enumName + " nameOf(String name) {\n"))
	_, _ = dst.Write([]byte("  	for (var item in values) {\n"))
	_, _ = dst.Write([]byte("  		if (item.name == name) {\n"))
	_, _ = dst.Write([]byte("  			return item;\n"))
	_, _ = dst.Write([]byte("  		}\n"))
	_, _ = dst.Write([]byte("  	}\n"))
	_, _ = dst.Write([]byte("  	throw 'Get " + enumName + " by name error, name=$name';\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	for _, item := range typ.Items {
		itemName := build.StringToAllUpper(item.Name.Name)
		_, _ = dst.Write([]byte("  static const " + itemName + " = " + enumName + "._(" + item.Id.Value + ", '" + itemName + "');\n"))
	}
	_, _ = dst.Write([]byte("\n"))
	_, _ = dst.Write([]byte("  static const List<" + enumName + "> values = [\n"))
	for _, item := range typ.Items {
		_, _ = dst.Write([]byte("    " + build.StringToAllUpper(item.Name.Name) + ",\n"))
	}
	_, _ = dst.Write([]byte("  ];\n\n"))

	_, _ = dst.Write([]byte("}\n\n"))
}
