package dart

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"io"
)

func printData(dst io.Writer, typ *ast.DataType) {
	_, _ = dst.Write([]byte("abstract class " + build.StringToHumpName(typ.Name.Name) + " implements Data"))
	if nil != typ.Extends {
		printExtend(dst, typ.Extends, true)
	}
	_, _ = dst.Write([]byte("{\n"))
	for _, field := range typ.Fields.List {
		if nil != field.Comment {
			_, _ = dst.Write([]byte("  /// " + field.Comment.Text()))
		}
		isSuper := build.CheckSuperField(field.Name.Name, typ)
		if isSuper {
			_, _ = dst.Write([]byte("  @override\n"))
		}
		_, _ = dst.Write([]byte("  "))
		printType(dst, field.Type, false)
		_, _ = dst.Write([]byte(" get " + build.StringToFirstLower(field.Name.Name)))
		_, _ = dst.Write([]byte(";\n\n"))

		if isSuper {
			_, _ = dst.Write([]byte("  @override\n"))
		}
		_, _ = dst.Write([]byte("  set "))
		_, _ = dst.Write([]byte(build.StringToFirstLower(field.Name.Name) + "("))
		printType(dst, field.Type, false)
		_, _ = dst.Write([]byte(" value);\n\n"))
	}
	isParam := false
	_, _ = dst.Write([]byte("  factory " + build.StringToHumpName(typ.Name.Name) + "("))
	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if !isParam {
			_, _ = dst.Write([]byte("{\n"))
			isParam = true
		}
		_, _ = dst.Write([]byte("    "))
		printType(dst, field.Type, true)
		_, _ = dst.Write([]byte(" " + build.StringToFirstLower(field.Name.Name)))
		_, _ = dst.Write([]byte(",\n"))
		return nil
	})
	if err != nil {
		return
	}
	if isParam {
		_, _ = dst.Write([]byte("}"))
	}
	_, _ = dst.Write([]byte("  ){\n"))
	_, _ = dst.Write([]byte("    return _" + build.StringToHumpName(typ.Name.Name) + "(\n"))
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		_, _ = dst.Write([]byte("      "))
		_, _ = dst.Write([]byte(build.StringToFirstLower(field.Name.Name)))
		_, _ = dst.Write([]byte(": "))
		_, _ = dst.Write([]byte(build.StringToFirstLower(field.Name.Name)))
		_, _ = dst.Write([]byte(",\n"))
		return nil
	})
	if err != nil {
		return
	}
	_, _ = dst.Write([]byte("    );\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("  static " + build.StringToHumpName(typ.Name.Name) + " fromMap(Map<String, dynamic> map){\n"))
	_, _ = dst.Write([]byte("    return _" + build.StringToHumpName(typ.Name.Name) + ".fromMap(map);\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("  static " + build.StringToHumpName(typ.Name.Name) + " fromData(ByteData data){\n"))
	_, _ = dst.Write([]byte("    return _" + build.StringToHumpName(typ.Name.Name) + ".fromData(data);\n"))
	_, _ = dst.Write([]byte("  }\n\n"))

	_, _ = dst.Write([]byte("}\n\n"))
}

func printDataEntity(dst io.Writer, typ *ast.DataType) {
	_, _ = dst.Write([]byte("class _" + build.StringToHumpName(typ.Name.Name) + " implements " + build.StringToHumpName(typ.Name.Name)))
	_, _ = dst.Write([]byte(" {\n"))

	err := build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		_, _ = dst.Write([]byte("  @override\n"))
		_, _ = dst.Write([]byte("  "))
		printType(dst, field.Type, false)
		_, _ = dst.Write([]byte(" " + build.StringToFirstLower(field.Name.Name)))
		_, _ = dst.Write([]byte(";\n\n"))
		return nil
	})
	if err != nil {
		return
	}

	_, _ = dst.Write([]byte("  _" + build.StringToHumpName(typ.Name.Name) + "("))
	isParam := false
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		if !isParam {
			_, _ = dst.Write([]byte("{\n"))
			isParam = true
		}
		_, _ = dst.Write([]byte("    "))
		if !field.Type.IsEmpty() {
			_, _ = dst.Write([]byte("required "))
		}
		_, _ = dst.Write([]byte("this." + build.StringToFirstLower(field.Name.Name)))
		_, _ = dst.Write([]byte(",\n"))
		return nil
	})
	if err != nil {
		return
	}
	_, _ = dst.Write([]byte("  "))
	if isParam {
		_, _ = dst.Write([]byte("}"))
	}
	_, _ = dst.Write([]byte(");\n\n"))

	_, _ = dst.Write([]byte("  static _" + build.StringToHumpName(typ.Name.Name) + " fromMap(Map<String, dynamic> map){\n"))
	_, _ = dst.Write([]byte("    return _" + build.StringToHumpName(typ.Name.Name) + "(\n"))

	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		_, _ = dst.Write([]byte("      " + build.StringToFirstLower(field.Name.Name)))
		_, _ = dst.Write([]byte(": map[\"" + getJsonName(field) + "\"]"))
		_, _ = dst.Write([]byte(",\n"))
		return nil
	})
	if err != nil {
		return
	}

	_, _ = dst.Write([]byte("    );\n"))
	_, _ = dst.Write([]byte("  }\n"))

	_, _ = dst.Write([]byte("\n"))
	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  Map<String, dynamic> toMap() {\n"))
	_, _ = dst.Write([]byte("    return {\n"))
	err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
		_, _ = dst.Write([]byte("      \"" + getJsonName(field)))
		_, _ = dst.Write([]byte("\": " + build.StringToFirstLower(field.Name.Name) + ",\n"))
		return nil
	})
	if err != nil {
		return
	}
	_, _ = dst.Write([]byte("    };\n"))
	_, _ = dst.Write([]byte("  }\n"))

	_, _ = dst.Write([]byte("  static _" + build.StringToHumpName(typ.Name.Name) + " fromData(ByteData data){\n"))
	_, _ = dst.Write([]byte("    return _" + build.StringToHumpName(typ.Name.Name) + "(\n"))

	//err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
	//	_, _ = dst.Write([]byte("      " + build.StringToFirstLower(field.Name.Name)))
	//	_, _ = dst.Write([]byte(": map[\"" + getJsonName(field) + "\"]"))
	//	_, _ = dst.Write([]byte(",\n"))
	//	return nil
	//})
	//if err != nil {
	//	return
	//}

	_, _ = dst.Write([]byte("    );\n"))
	_, _ = dst.Write([]byte("  }\n"))

	_, _ = dst.Write([]byte("\n"))
	_, _ = dst.Write([]byte("  @override\n"))
	_, _ = dst.Write([]byte("  ByteData toData() {\n"))
	_, _ = dst.Write([]byte("    return ByteData.view(Uint8List(12).buffer);\n"))
	//err = build.EnumField(typ, func(field *ast.Field, data *ast.DataType) error {
	//	_, _ = dst.Write([]byte("      \"" + getJsonName(field)))
	//	_, _ = dst.Write([]byte("\": " + build.StringToFirstLower(field.Name.Name) + ",\n"))
	//	return nil
	//})
	//if err != nil {
	//	return
	//}
	//_, _ = dst.Write([]byte("    };\n"))
	_, _ = dst.Write([]byte("  }\n"))

	_, _ = dst.Write([]byte("}\n\n"))
}

func printExtend(dst io.Writer, extends []*ast.Ident, start bool) {
	for i, v := range extends {
		if 0 != i || start {
			_, _ = dst.Write([]byte(", "))
		}
		_, _ = dst.Write([]byte(build.StringToHumpName(v.Name)))

	}
}
