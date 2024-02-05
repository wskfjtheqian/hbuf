package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printDataCode(dst *build.Writer, typ *ast.DataType) {
	dst.Import("encoding/json", "")
	name := build.StringToHumpName(typ.Name.Name)
	if nil != typ.Doc && 0 < len(typ.Doc.Text()) {
		dst.Code("// " + name + " " + typ.Doc.Text())
	}
	dst.Code("type " + name + " struct")
	dst.Code(" {\n")
	isFast := true
	b.printDataExtend(dst, typ.Extends, &isFast)

	for _, field := range typ.Fields.List {
		if !isFast {
			dst.Code("\n")
		}
		isFast = false
		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Code("\t//" + field.Doc.Text())
		}

		dst.Code("\t" + build.StringToHumpName(field.Name.Name) + " ")
		b.printType(dst, field.Type, true)

		dst.Code(" `json:\"" + build.StringToUnderlineName(field.Name.Name) + "\"`")
		dst.Code("\n")
	}
	dst.Code("}\n\n")

	dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") ToData() ([]byte, error) {\n")
	dst.Code("\treturn json.Marshal(g)\n")
	dst.Code("}\n\n")

	dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") FormData(data []byte) error {\n")
	dst.Code("\treturn json.Unmarshal(data, g)\n")
	dst.Code("}\n\n")

	for _, field := range typ.Fields.List {
		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Code("// Get" + build.StringToHumpName(field.Name.Name) + " Get " + field.Doc.Text())
		}
		dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") Get" + build.StringToHumpName(field.Name.Name) + "() ")
		b.printType(dst, field.Type, false)
		dst.Code(" {\n")
		if field.Type.IsEmpty() && !build.IsArray(field.Type) && !build.IsMap(field.Type) {
			dst.Code("\tif nil == g." + build.StringToHumpName(field.Name.Name) + " {\n")
			dst.Code("\t\treturn ")
			b.printDefault(dst, field.Type)
			dst.Code("\n")
			dst.Code("\t}\n")
			dst.Code("\treturn *g." + build.StringToHumpName(field.Name.Name) + "\n")
		} else {
			dst.Code("\treturn g." + build.StringToHumpName(field.Name.Name) + "\n")
		}
		dst.Code("}\n\n")

		if nil != field.Doc && 0 < len(field.Doc.Text()) {
			dst.Code("// Set" + build.StringToHumpName(field.Name.Name) + " Set " + field.Doc.Text())
		}
		dst.Code("func (g *" + build.StringToHumpName(typ.Name.Name) + ") Set" + build.StringToHumpName(field.Name.Name) + "(val ")
		b.printType(dst, field.Type, false)
		dst.Code(") {\n")
		dst.Code("\tg." + build.StringToHumpName(field.Name.Name) + " = ")
		if field.Type.IsEmpty() && !build.IsArray(field.Type) && !build.IsMap(field.Type) {
			dst.Code("&val\n")
		} else {
			dst.Code("val\n")
		}
		dst.Code("}\n\n")
	}
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
			dst.Code(t.DefaultValue())
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
		dst.Code("\t")
		pack := b.getPackage(dst, v.Name)
		dst.Code(pack)
		dst.Code(build.StringToHumpName(v.Name.Name))
		dst.Code("\n")
	}
}
