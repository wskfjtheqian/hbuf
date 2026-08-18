package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"sort"
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
	dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf")
	dst.Import("reflect")
	dst.Import("unsafe")

	dst.Code("var ").Code(name).Code("Descriptor = hbuf.NewSyncDescriptor(func() hbuf.Descriptor {\n")
	dst.Tab(1).Code("var ").Code(name).Code(" ").Code(build.StringToHumpName(typ.Name.Name)).Code("\n")
	dst.Tab(1).Code("return hbuf.NewDataDescriptor(0, false, reflect.TypeOf(&").Code(name).Code("), map[uint16]hbuf.Descriptor{")

	if len(typ.Extends) > 0 {
		dst.Code("\n")
	}
	id := 0
	for _, extend := range typ.Extends {
		v, _ := strconv.Atoi(extend.Id.Value)
		if id < v {
			id = v
		}
	}
	length := len(strconv.Itoa(id)) + 1
	for _, extend := range typ.Extends {
		dst.Tab(2).Code(extend.Id.Value).Code(":").Code(strings.Repeat(" ", length-len(extend.Id.Value)))
		b.printDescriptor(dst, extend.Name, false, name, build.StringToHumpName(extend.Name.Name))
		dst.Code(",\n")
	}
	dst.Code("}, map[uint16]hbuf.Descriptor{")
	if len(typ.Fields.List) > 0 {
		dst.Code("\n")
	}

	id = 0
	for _, field := range typ.Fields.List {
		v, _ := strconv.Atoi(field.Id.Value)
		if id < v {
			id = v
		}
	}
	length = len(strconv.Itoa(id)) + 1
	for _, field := range typ.Fields.List {
		dst.Tab(2).Code(field.Id.Value).Code(":").Code(strings.Repeat(" ", length-len(field.Id.Value)))
		b.printDescriptor(dst, field.Type, false, name, build.StringToHumpName(field.Name.Name))
		dst.Code(",\n")
	}
	dst.Tab(1).Code("})\n")
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
		dst.Code(", ").Code(isPrt).Code(")")

	case *ast.MapType:
		ma := expr.(*ast.MapType)

		dst.Code("hbuf.NewMapDescriptor[").Code(b.getDescriptorType(dst, ma.Key, false)).Code(", ").Code(b.getDescriptorType(dst, ma.VType, false)).Code("](").Code(offsetof).Code(", ")
		b.printDescriptor(dst, ma.Key, false, "", "")
		dst.Code(", ")
		b.printDescriptor(dst, ma.VType, false, "", "")
		dst.Code(", ").Code(isPrt).Code(")")
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printDescriptor(dst, t.Type(), isNull || t.Empty, structName, fieldName)
	}
}

func (b *Builder) printDataStruct(dst *build.Writer, typ *ast.DataType) error {
	db, _, _, err := b.getDBField(typ)
	if err != nil {
		return err
	}
	uName := build.StringToHumpName(typ.Name.Name)
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("// ").Code(uName).Code(" ").Code(typ.Doc.Text())
	}
	dst.Code("type ").Code(uName).Code(" struct {\n")

	length := 0
	nameLen := 0
	typLen := 0
	tagLen := 0

	isChange := false
	if db != nil {
		val := strings.ToLower(db.Change)
		if "self" == val || "parent" == val {
			isChange = true
		}
	}

	fields := make([]dataField, 0, len(typ.Fields.List)+1)
	if isChange {
		nameLen = len("changeFields")
		typLen = len("[]bool")
		fields = append(fields, dataField{
			name: "changeFields",
			typ:  "[]bool",
		})
	}

	for _, field := range typ.Fields.List {
		temp := build.NewWriter()
		temp.Packages = dst.Packages
		b.printType(temp, field.Type, true)
		dst.AddImports(temp.GetImports())

		marshal := build.GetMarshal(field.Tags)

		inOut := ""

		if marshal != nil {
			for i, in := range marshal.In {
				if len(in) > 0 {
					if i > 0 {
						inOut += "|"
					}
					inOut += "I" + in
				}
			}
			for i, out := range marshal.Out {
				if len(out) > 0 {
					if i > 0 {
						inOut += "|"
					}
					inOut += "O" + out
				}
			}
			inOut = ",filter:" + inOut
		}

		tag := strings.Builder{}
		tag.WriteString("`")
		tags := build.GetFieldTag(field.Tags)
		if len(tags) > 0 {
			if _, ok := tags["json"]; !ok {
				tag.WriteString("json:\"" + build.StringToUnderlineName(field.Name.Name) + ",omitempty" + inOut + "\"")
			}
			keys := make([]string, 0)
			for key, _ := range tags {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			for _, key := range keys {
				if val, ok := tags[key]; ok {
					if tag.Len() > 0 {
						tag.WriteString(" ")
					}
					tag.WriteString(key + ":\"" + strings.Join(val, ",") + "\"")
				}
			}
		} else {
			tag.WriteString("json:\"" + build.StringToUnderlineName(field.Name.Name) + ",omitempty\"")
		}

		tag.WriteString(" hbuf:\"" + field.Id.Value + "\"")
		tag.WriteString("`")

		fields = append(fields, dataField{
			name: build.StringToHumpName(field.Name.Name),
			typ:  temp.String(),
			tag:  tag.String(),
		})
		i := len(fields) - 1

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
		dst.Tab(1)
		dst.Code(build.StringFillRight(field.name, ' ', nameLen+1))
		dst.Code(build.StringFillRight(field.typ, ' ', typLen+1))
		dst.Code(build.StringFillRight(field.tag, ' ', tagLen))
		dst.Code("//").Code(strings.Trim(strings.ReplaceAll(field.comment, "\n", " "), " ")).Code("\n")
	}
	dst.Code("}\n\n")

	dst.Code("func (g *").Code(uName).Code(") Descriptors() hbuf.Descriptor {\n")
	dst.Tab(1).Code("return ").Code(build.StringToFirstLower(typ.Name.Name)).Code("Descriptor.Desc()\n")
	dst.Code("}\n\n")

	for _, field := range typ.Fields.List {
		uFieldName := build.StringToHumpName(field.Name.Name)
		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Code("// Get").Code(uFieldName).Code(" Get ").Code(field.Doc.Text())
		}
		dst.Code("func (g *").Code(uName).Code(") Get").Code(uFieldName).Code("() ")
		b.printType(dst, field.Type, false)
		dst.Code(" {\n")
		if field.Type.IsEmpty() && !build.IsArray(field.Type) && !build.IsMap(field.Type) {
			dst.Tab(1).Code("if nil == g.").Code(uFieldName).Code(" {\n")
			dst.Tab(2).Code("return ")
			b.printDefault(dst, field.Type)
			dst.Code("\n")
			dst.Tab(1).Code("}\n")
			dst.Tab(1).Code("return *g.").Code(uFieldName).Code("\n")
		} else {
			dst.Tab(1).Code("return g.").Code(uFieldName).Code("\n")
		}
		dst.Code("}\n\n")

		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Code("// Set").Code(uFieldName).Code(" Set ").Code(field.Doc.Text())
		}
		dst.Code("func (g *").Code(uName).Code(") Set").Code(uFieldName).Code("(val ")
		b.printType(dst, field.Type, false)
		dst.Code(") {\n")
		dst.Tab(1).Code("g.").Code(uFieldName).Code(" = ")
		if field.Type.IsEmpty() && !build.IsArray(field.Type) && !build.IsMap(field.Type) {
			dst.Code("&val\n")
		} else {
			dst.Code("val\n")
		}
		if isChange {
			temp := build.GetDB(field.Name.Name, field.Tags)
			if temp != nil {
				dst.Tab(1).Code("g.changeFields[int(").Code(uName).Code("Field_").Code(uFieldName).Code(")] = true\n")
			}
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
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf")
			} else if build.Decimal == t {
				dst.Import("github.com/shopspring/decimal")
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
