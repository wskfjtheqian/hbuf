package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strings"
)

type dataField struct {
	name    string
	typ     string
	tag     string
	comment string
}

func (b *Builder) printDataCode(dst *build.Writer, typ *ast.DataType) error {
	err := b.printDataDescriptor(dst, typ)
	if err != nil {
		return err
	}

	err = b.printDataStruct(dst, typ)
	if err != nil {
		return err
	}

	return nil
}

func (b *Builder) printDataDescriptor(dst *build.Writer, typ *ast.DataType) error {
	name := build.StringToHumpName(typ.Name.Name)
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")

	dst.Code("var ").Code(name).Code("Fields = map[uint16]hbuf.Descriptor{\n")
	for _, field := range typ.Fields.List {
		dst.Tab(1).Code(field.Id.Value).Code(":")
		b.printDescriptor(dst, field.Type, true, name, build.StringToHumpName(field.Name.Name))
		dst.Code(",\n")
	}
	dst.Code("}\n\n")

	return nil
}

func (b *Builder) printDescriptor(dst *build.Writer, expr ast.Expr, b2 bool, structName string, fieldName string) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			pack := b.getPackage(dst, expr)
			dst.Code(pack + (expr.(*ast.Ident)).Name)
		} else {
			switch build.BaseType((expr.(*ast.Ident)).Name) {
			case build.Int8:
				dst.Code("hbuf.NewInt64Descriptor(func(d any) int64 {\n")
				dst.Tab(2).Code("return int64(d.(*").Code(structName).Code(").").Code(fieldName).Code(")\n")
				dst.Tab(1).Code("}, func(d any, v int64) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = int8(v)\n")
				dst.Tab(1).Code("})")
			case build.Int16:
				dst.Code("hbuf.NewInt16Descriptor(func(d any) int16 {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v int16) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			case build.Int32:
				dst.Code("hbuf.NewInt32Descriptor(func(d any) int32 {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v int32) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			case build.Int64:
				dst.Code("hbuf.NewInt64Descriptor(func(d any) hbuf.Int64 {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v hbuf.Int64) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			case build.Uint8:
				dst.Code("hbuf.NewUint8Descriptor(func(d any) uint8 {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v uint8) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			case build.Uint16:
				dst.Code("hbuf.NewUint16Descriptor(func(d any) uint16 {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v uint16) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			case build.Uint32:
				dst.Code("hbuf.NewUint32Descriptor(func(d any) uint32 {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v uint32) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			case build.Uint64:
				dst.Code("hbuf.NewUint64Descriptor(func(d any) hbuf.Uint64 {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v hbuf.Uint64) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			case build.Float:
				dst.Code("hbuf.NewFloatDescriptor(func(d any) float32 {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v float32) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			case build.Double:
				dst.Code("hbuf.NewDoubleDescriptor(func(d any) float64 {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v float64) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			case build.Bool:
				dst.Code("hbuf.NewBoolDescriptor(func(d any) bool {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v bool) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			case build.String:
				dst.Code("hbuf.NewStringDescriptor(func(d any) string {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v string) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			case build.Decimal:
				dst.Code("hbuf.NewDecimalDescriptor(func(d any) decimal.Decimal {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v decimal.Decimal) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			case build.Date:
				dst.Code("hbuf.NewTimeDescriptor(func(d any) hbuf.Time {\n")
				dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
				dst.Tab(1).Code("}, func(d any, v hbuf.Time) {\n")
				dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
				dst.Tab(1).Code("})")
			}
		}
	case *ast.ArrayType:
		//dst.Code("hbuf.NewListDescriptor(func(d any) any {\n")
		//dst.Tab(2).Code("return d.(*").Code(structName).Code(").").Code(fieldName).Code("\n")
		//dst.Tab(1).Code("}, func(d any, v any) {\n")
		//dst.Tab(2).Code("d.(*").Code(structName).Code(").").Code(fieldName).Code(" = v\n")
		//dst.Tab(1).Code("}), ")
		//b.printDescriptor(dst, expr.(*ast.ArrayType).VType, true, structName, fieldName)

		//ar := expr.(*ast.ArrayType)
		//dst.Code("[]")
		//b.printDescriptor(dst, ar.VType, true)
	case *ast.MapType:
		//ma := expr.(*ast.MapType)
		//dst.Code("map[")
		//b.printDescriptor(dst, ma.Key, true)
		//dst.Code("]")
		//b.printDescriptor(dst, ma.VType, true)
	case *ast.VarType:
		t := expr.(*ast.VarType)
		//if b2 && t.Empty {
		//	dst.Code("*")
		//}
		b.printDescriptor(dst, t.Type(), true, structName, fieldName)
	}
}

func (b *Builder) printDataStruct(dst *build.Writer, typ *ast.DataType) error {
	dst.Import("encoding/json", "")
	name := build.StringToHumpName(typ.Name.Name)
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("// " + name + " " + typ.Doc.Text())
	}
	dst.Code("type " + name + " struct")
	dst.Code(" {\n")

	length := 0
	nameLen := 0
	typLen := 0
	tagLen := 0
	fields := make([]dataField, len(typ.Fields.List))
	for i, field := range typ.Fields.List {
		temp := build.NewWriter()
		temp.Packages = dst.Packages
		b.printType(temp, field.Type, true)
		dst.AddImports(temp.GetImports())

		fields[i] = dataField{
			name: build.StringToHumpName(field.Name.Name),
			typ:  temp.String(),
			tag:  "`json:\"" + build.StringToUnderlineName(field.Name.Name) + ",omitempty\" hbuf:\"" + field.Id.Value + "\"` ",
		}

		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			fields[i].comment = field.Doc.Text()
		}

		length = len(fields[i].name)
		if length > nameLen {
			nameLen = length
		}
		length = len(fields[i].typ)
		if length > typLen {
			typLen = length
		}
		length = len(fields[i].tag)
		if length > tagLen {
			tagLen = length
		}
	}

	isFast := true
	b.printDataExtend(dst, typ.Extends, &isFast)
	for _, field := range fields {
		dst.Tab(1).Code("")
		dst.Code(build.StringFillRight(field.name, ' ', nameLen+1))
		dst.Code(build.StringFillRight(field.typ, ' ', typLen+1))
		dst.Code(build.StringFillRight(field.tag, ' ', tagLen+1))
		dst.Code("//").Code(strings.Trim(strings.ReplaceAll(field.comment, "\n", " "), " ")).Code("\n")
	}
	dst.Code("}\n\n")

	dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") ToData() ([]byte, error) {\n")
	dst.Tab(1).Code("return json.Marshal(g)\n")
	dst.Code("}\n\n")

	dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") FormData(data []byte) error {\n")
	dst.Tab(1).Code("return json.Unmarshal(data, g)\n")
	dst.Code("}\n\n")

	for _, field := range typ.Fields.List {
		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Code("// Get" + build.StringToHumpName(field.Name.Name) + " Get " + field.Doc.Text())
		}
		dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") Get" + build.StringToHumpName(field.Name.Name) + "() ")
		b.printType(dst, field.Type, false)
		dst.Code(" {\n")
		if field.Type.IsEmpty() && !build.IsArray(field.Type) && !build.IsMap(field.Type) {
			dst.Tab(1).Code("if nil == g." + build.StringToHumpName(field.Name.Name) + " {\n")
			dst.Tab(2).Code("return ")
			b.printDefault(dst, field.Type)
			dst.Code("\n")
			dst.Tab(1).Code("}\n")
			dst.Tab(1).Code("return *g." + build.StringToHumpName(field.Name.Name) + "\n")
		} else {
			dst.Tab(1).Code("return g." + build.StringToHumpName(field.Name.Name) + "\n")
		}
		dst.Code("}\n\n")

		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Code("// Set" + build.StringToHumpName(field.Name.Name) + " Set " + field.Doc.Text())
		}
		dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") Set" + build.StringToHumpName(field.Name.Name) + "(val ")
		b.printType(dst, field.Type, false)
		dst.Code(") {\n")
		dst.Tab(1).Code("g." + build.StringToHumpName(field.Name.Name) + " = ")
		if field.Type.IsEmpty() && !build.IsArray(field.Type) && !build.IsMap(field.Type) {
			dst.Code("&val\n")
		} else {
			dst.Code("val\n")
		}
		dst.Code("}\n\n")
	}
	return nil
}

func (b *Builder) printDefault(dst *build.Writer, expr ast.Expr) {
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			pack := b.getPackage(dst, expr)
			dst.Code(pack + (expr.(*ast.Ident)).Name)
			if t.Obj.Kind == ast.Enum {
				dst.Code("(0)")
			} else {
				dst.Code("{}")
			}
		} else {
			t := build.BaseType((expr.(*ast.Ident)).Name)
			if build.Date == t || build.Uint64 == t || build.Int64 == t {
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
			} else if build.Decimal == t {
				dst.Import("github.com/shopspring/decimal", "")
			}
			if val, ok := _typesDefaultValue[t]; ok {
				dst.Code(val)
			} else {
				dst.Code("")
			}
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		dst.Code("[]")
		b.printType(dst, ar.VType, true)
		dst.Code("{}")
	case *ast.MapType:
		ma := expr.(*ast.MapType)
		dst.Code("map[")
		b.printType(dst, ma.Key, true)
		dst.Code("]")
		b.printType(dst, ma.VType, true)
		dst.Code("{}")
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printDefault(dst, t.Type())
	}
}

func (b *Builder) printDataExtend(dst *build.Writer, extends []*ast.Extends, isFast *bool) {
	for _, v := range extends {
		if !*isFast {
			dst.Code("\n")
		}
		*isFast = false
		dst.Tab(1).Code("")
		pack := b.getPackage(dst, v.Name)
		dst.Code(pack)
		dst.Code(build.StringToHumpName(v.Name.Name))
		dst.Code("\n")
	}

}
