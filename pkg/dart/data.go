package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func printDataCode(dst *Writer, typ *ast.DataType) {
	dst.Import("dart:typed_data")
	dst.Import("package:hbuf_dart/hbuf_dart.dart")

	printData(dst, typ)
	printDataEntity(dst, typ)
}
func printData(dst *Writer, typ *ast.DataType) {
	dst.Code("abstract class " + build.StringToHumpName(typ.Name.Name) + " implements Data")
	if nil != typ.Extends {
		printExtend(dst, typ.Extends, true)
	}
	dst.Code("{\n")
	for _, field := range typ.Fields.List {
		if nil != field.Comment {
			dst.Code("  /// " + field.Comment.Text())
		}
		isSuper := build.CheckSuperField(field.Name.Name, typ)
		if isSuper {
			dst.Code("  @override\n")
		}
		dst.Code("  ")
		printType(dst, field.Type, false)
		dst.Code(" get " + build.StringToFirstLower(field.Name.Name))
		dst.Code(";\n\n")

		if isSuper {
			dst.Code("  @override\n")
		}
		dst.Code("  set ")
		dst.Code(build.StringToFirstLower(field.Name.Name) + "(")
		printType(dst, field.Type, false)
		dst.Code(" value);\n\n")
	}
	isParam := false
	dst.Code("  factory " + build.StringToHumpName(typ.Name.Name) + "(")
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if !isParam {
			dst.Code("{\n")
			isParam = true
		}
		dst.Code("    ")
		printType(dst, field.Type, true)
		dst.Code(" " + build.StringToFirstLower(field.Name.Name))
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	if isParam {
		dst.Code("}")
	}
	dst.Code("  ){\n")
	dst.Code("    return _" + build.StringToHumpName(typ.Name.Name) + "(\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("      ")
		dst.Code(build.StringToFirstLower(field.Name.Name))
		dst.Code(": ")
		dst.Code(build.StringToFirstLower(field.Name.Name))
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("    );\n")
	dst.Code("  }\n\n")

	dst.Code("  static " + build.StringToHumpName(typ.Name.Name) + " fromMap(Map<String, dynamic> map){\n")
	dst.Code("    return _" + build.StringToHumpName(typ.Name.Name) + ".fromMap(map);\n")
	dst.Code("  }\n\n")

	dst.Code("  static " + build.StringToHumpName(typ.Name.Name) + " fromData(ByteData data){\n")
	dst.Code("    return _" + build.StringToHumpName(typ.Name.Name) + ".fromData(data);\n")
	dst.Code("  }\n\n")

	dst.Code("}\n\n")
}

func printDataEntity(dst *Writer, typ *ast.DataType) {
	dst.Code("class _" + build.StringToHumpName(typ.Name.Name) + " implements " + build.StringToHumpName(typ.Name.Name))
	dst.Code(" {\n")

	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("  @override\n")
		dst.Code("  ")
		printType(dst, field.Type, false)
		dst.Code(" " + build.StringToFirstLower(field.Name.Name))
		dst.Code(";\n\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Code("  _" + build.StringToHumpName(typ.Name.Name) + "(")
	isParam := false
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if !isParam {
			dst.Code("{\n")
			isParam = true
		}
		dst.Code("    ")
		if !field.Type.IsEmpty() {
			dst.Code("required ")
		}
		dst.Code("this." + build.StringToFirstLower(field.Name.Name))
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("  ")
	if isParam {
		dst.Code("}")
	}
	dst.Code(");\n\n")

	dst.Code("  static _" + build.StringToHumpName(typ.Name.Name) + " fromMap(Map<String, dynamic> map){\n")
	dst.Code("    return _" + build.StringToHumpName(typ.Name.Name) + "(\n")

	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("      " + build.StringToFirstLower(field.Name.Name))
		dst.Code(": map[\"" + getJsonName(field) + "\"]")
		dst.Code(",\n")
		return nil
	})
	if err != nil {
		return
	}

	dst.Code("    );\n")
	dst.Code("  }\n")

	dst.Code("\n")
	dst.Code("  @override\n")
	dst.Code("  Map<String, dynamic> toMap() {\n")
	dst.Code("    return {\n")
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		dst.Code("      \"" + getJsonName(field))
		dst.Code("\": " + build.StringToFirstLower(field.Name.Name) + ",\n")
		return nil
	})
	if err != nil {
		return
	}
	dst.Code("    };\n")
	dst.Code("  }\n")

	dst.Code("  static _" + build.StringToHumpName(typ.Name.Name) + " fromData(ByteData data){\n")
	dst.Code("    return _" + build.StringToHumpName(typ.Name.Name) + ".fromMap({});\n")

	//err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
	//	dst.Code("      " + build.StringToFirstLower(field.Name.Name))
	//	dst.Code(": map[\"" + getJsonName(field) + "\"]")
	//	dst.Code(",\n")
	//	return nil
	//})
	//if err != nil {
	//	return
	//}

	//dst.Code("    );\n")
	dst.Code("  }\n")

	dst.Code("\n")
	dst.Code("  @override\n")
	dst.Code("  ByteData toData() {\n")
	dst.Code("    return ByteData.view(Uint8List(12).buffer);\n")
	//err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
	//	dst.Code("      \"" + getJsonName(field))
	//	dst.Code("\": " + build.StringToFirstLower(field.Name.Name) + ",\n")
	//	return nil
	//})
	//if err != nil {
	//	return
	//}
	//dst.Code("    };\n")
	dst.Code("  }\n")

	dst.Code("}\n\n")
}

func printExtend(dst *Writer, extends []*ast.Ident, start bool) {
	for i, v := range extends {
		if 0 != i || start {
			dst.Code(", ")
		}
		dst.Code(build.StringToHumpName(v.Name))

	}
}
