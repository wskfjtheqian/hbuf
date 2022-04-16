package dart

import (
	"hbuf/pkg/ast"
	"io"
)

func printEnum(dst io.Writer, typ *ast.EnumType) {
	_, _ = dst.Write([]byte("class " + typ.Name.Name))
	_, _ = dst.Write([]byte("{\n"))
	_, _ = dst.Write([]byte("  final int value;\n"))
	_, _ = dst.Write([]byte("  final String name;\n\n"))

	_, _ = dst.Write([]byte("  const " + typ.Name.Name + "._(this.value, this.name);\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  bool operator ==(Object other) =>\n"))
	_, _ = dst.Write([]byte("      identical(this, other) ||\n"))
	_, _ = dst.Write([]byte("      other is Gender &&\n"))
	_, _ = dst.Write([]byte("          runtimeType == other.runtimeType &&\n"))
	_, _ = dst.Write([]byte("          value == other.value;\n\n"))

	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  int get hashCode => value.hashCode;\n\n"))

	_, _ = dst.Write([]byte("  static Gender valueOf(int value) {\n"))
	_, _ = dst.Write([]byte("  	for (var item in values) {\n"))
	_, _ = dst.Write([]byte("  		if (item.value == value) {\n"))
	_, _ = dst.Write([]byte("  			return item;\n"))
	_, _ = dst.Write([]byte("  		}\n"))
	_, _ = dst.Write([]byte("  	}\n"))
	_, _ = dst.Write([]byte("  	throw 'Get Gender by value error, value=$value';\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("  static Gender nameOf(String name) {\n"))
	_, _ = dst.Write([]byte("  	for (var item in values) {\n"))
	_, _ = dst.Write([]byte("  		if (item.name == name) {\n"))
	_, _ = dst.Write([]byte("  			return item;\n"))
	_, _ = dst.Write([]byte("  		}\n"))
	_, _ = dst.Write([]byte("  	}\n"))
	_, _ = dst.Write([]byte("  	throw 'Get Gender by name error, name=$name';\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	for _, item := range typ.Items {
		_, _ = dst.Write([]byte("  static const " + item.Name.Name + " = Gender._(" + item.Id.Value + ", '" + item.Name.Name + "');\n"))
	}
	_, _ = dst.Write([]byte("\n"))
	_, _ = dst.Write([]byte("  static const List<Gender> values = [\n"))
	for _, item := range typ.Items {
		_, _ = dst.Write([]byte("    " + item.Name.Name + ",\n"))
	}
	_, _ = dst.Write([]byte("  ];\n\n"))

	_, _ = dst.Write([]byte("}\n\n"))
}
