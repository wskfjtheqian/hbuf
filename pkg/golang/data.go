package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strconv"
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

	name := build.StringToFirstLower(typ.Name.Name)
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
	dst.Import("reflect", "")
	dst.Import("unsafe", "")

	dst.Code("var ").Code(name).Code(" ").Code(build.StringToHumpName(typ.Name.Name)).Code("\n")
	dst.Code("var ").Code(name).Code("Descriptor = hbuf.NewDataDescriptor(0, false, reflect.TypeOf(&").Code(name).Code("), map[uint16]hbuf.Descriptor{\n")

	id := 0
	for _, extend := range typ.Extends {
		v, _ := strconv.Atoi(extend.Id.Value)
		if id < v {
			id = v
		}
	}
	length := len(strconv.Itoa(id)) + 1
	for _, extend := range typ.Extends {
		dst.Tab(1).Code(extend.Id.Value).Code(":").Code(strings.Repeat(" ", length-len(extend.Id.Value)))
		b.printDescriptor(dst, extend.Name, false, name, build.StringToHumpName(extend.Name.Name))
		dst.Code(",\n")
	}
	dst.Code("}, map[uint16]hbuf.Descriptor{\n")

	id = 0
	for _, field := range typ.Fields.List {
		v, _ := strconv.Atoi(field.Id.Value)
		if id < v {
			id = v
		}
	}
	length = len(strconv.Itoa(id)) + 1
	for _, field := range typ.Fields.List {
		dst.Tab(1).Code(field.Id.Value).Code(":").Code(strings.Repeat(" ", length-len(field.Id.Value)))
		b.printDescriptor(dst, field.Type, false, name, build.StringToHumpName(field.Name.Name))
		dst.Code(",\n")
	}
	dst.Code("})\n\n")

	return nil
}
func (b *Builder) getDescriptorType(dst *build.Writer, expr ast.Expr, isNull bool) string {
	isPrt := ""
	if isNull {
		isPrt = "*"
	}
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			pack := b.getPackage(dst, expr)
			return isPrt + pack + build.StringToHumpName((expr.(*ast.Ident)).Name)
		} else {
			return isPrt + _types[build.BaseType((expr.(*ast.Ident)).Name)]
		}
	case *ast.VarType:
		t := expr.(*ast.VarType)
		return b.getDescriptorType(dst, t.Type(), t.Empty)
	}
	return ""
}

func (b *Builder) printDescriptor(dst *build.Writer, expr ast.Expr, isNull bool, structName string, fieldName string) {
	isPrt := "false"
	if isNull {
		isPrt = "true"
	}

	offsetof := "0"
	if structName != "" && fieldName != "" {
		offsetof = "unsafe.Offsetof(" + structName + "." + fieldName + ")"
	}
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			pack := b.getPackage(dst, expr)
			if ast.Enum == t.Obj.Kind {
				dst.Code("hbuf.NewInt32Descriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			} else {
				dst.Code("hbuf.CloneDataDescriptor(&").Code(pack + build.StringToHumpName((expr.(*ast.Ident)).Name)).Code("{}, ").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			}
		} else {
			switch build.BaseType((expr.(*ast.Ident)).Name) {
			case build.Int8:
				dst.Code("hbuf.NewInt8Descriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Int16:
				dst.Code("hbuf.NewInt16Descriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Int32:
				dst.Code("hbuf.NewInt32Descriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Int64:
				dst.Code("hbuf.NewInt64Descriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Uint8:
				dst.Code("hbuf.NewUint8Descriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Uint16:
				dst.Code("hbuf.NewUint16Descriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Uint32:
				dst.Code("hbuf.NewUint32Descriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Uint64:
				dst.Code("hbuf.NewUint64Descriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Float:
				dst.Code("hbuf.NewFloatDescriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Double:
				dst.Code("hbuf.NewDoubleDescriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Bool:
				dst.Code("hbuf.NewBoolDescriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.String:
				dst.Code("hbuf.NewStringDescriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Decimal:
				dst.Code("hbuf.NewDecimalDescriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Date:
				dst.Code("hbuf.NewTimeDescriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			case build.Bytes:
				dst.Code("hbuf.NewBytesDescriptor(").Code(offsetof).Code(", ").Code(isPrt).Code(")")
			}
		}
	case *ast.ArrayType:
		dst.Code("hbuf.NewListDescriptor[").Code(b.getDescriptorType(dst, expr.(*ast.ArrayType).VType, false)).Code("](").Code(offsetof).Code(", ")
		b.printDescriptor(dst, expr.(*ast.ArrayType).VType, false, "", "")
		dst.Code(",").Code(isPrt).Code(")")

	case *ast.MapType:
		ma := expr.(*ast.MapType)

		dst.Code("hbuf.NewMapDescriptor[").Code(b.getDescriptorType(dst, ma.Key, false)).Code(",").Code(b.getDescriptorType(dst, ma.VType, false)).Code("](").Code(offsetof).Code(", ")
		b.printDescriptor(dst, ma.Key, false, "", "")
		dst.Code(", ")
		b.printDescriptor(dst, ma.VType, false, "", "")
		dst.Code(",").Code(isPrt).Code(")")
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printDescriptor(dst, t.Type(), isNull || t.Empty, structName, fieldName)
	}
}

func (b *Builder) printDataStruct(dst *build.Writer, typ *ast.DataType) error {
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

	dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") Descriptors() hbuf.Descriptor  {\n")
	dst.Tab(1).Code("return ").Code(build.StringToFirstLower(typ.Name.Name)).Code("Descriptor\n")
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
	length := 0
	for _, extend := range extends {
		if length < len(extend.Name.Name) {
			length = len(extend.Name.Name)
		}
	}

	for _, v := range extends {
		*isFast = false
		dst.Tab(1).Code("")
		pack := b.getPackage(dst, v.Name)
		dst.Code(pack)
		dst.Code(build.StringToHumpName(v.Name.Name)).Code(strings.Repeat(" ", length-len(v.Name.Name)))
		dst.Code(" `hbuf:\"").Code(v.Id.Value).Code("\"`\n")
	}
	if len(extends) > 0 {
		dst.Code("\n")
	}
}
