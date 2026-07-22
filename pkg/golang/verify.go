package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strconv"
	"time"
)

func (b *Builder) printVerifyCode(dst *build.Writer, data *ast.DataType) error {
	dst.Import("context", "", 0)
	err := b.printVerifyFieldCode(dst, data)
	if err != nil {
		return err
	}

	verify, err := build.GetVerify(data.Tags, dst.File, b.GetDataType)
	if err != nil {
		return err
	}
	if nil == verify {
		return nil
	}

	err = b.printVerifyDataCode(dst, data)
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) GetDataType(file *ast.File, name string) *ast.Object {
	if obj := file.Scope.Lookup(name); nil != obj {
		switch obj.Decl.(type) {
		case *ast.TypeSpec:
			t := (obj.Decl.(*ast.TypeSpec)).Type
			switch t.(type) {
			case *ast.DataType:
				return obj
			case *ast.EnumType:
				return obj
			}
		}
	}
	for _, spec := range file.Imports {
		if f, ok := b.pkg.Files[spec.Path.Value]; ok {
			if obj := f.Scope.Lookup(name); nil != obj {
				switch obj.Decl.(type) {
				case *ast.TypeSpec:
					t := (obj.Decl.(*ast.TypeSpec)).Type
					switch t.(type) {
					case *ast.DataType:
						return obj
					case *ast.EnumType:
						return obj
					case *ast.ServerType:
						return obj
					}
				}
			}
		}
	}
	return nil
}

func (b *Builder) printVerifyFieldCode(dst *build.Writer, data *ast.DataType) error {
	dName := build.StringToHumpName(data.Name.Name)
	err := build.EnumField(data, func(field *ast.Field, data *ast.DataType) error {
		fName := build.StringToHumpName(field.Name.Name)

		verify, err := build.GetVerify(field.Tags, dst.File, b.GetDataType)
		if err != nil {
			return err
		}
		if nil == verify {
			return nil
		}

		dst.Code("func (i *" + dName + ") Verify" + fName + "(ctx context.Context, fields ...string) error {\n")

		first := true
		for i, val := range verify.GetFormat() {
			f := build.GetFormat(val.Item.Tags)
			if nil == f {
				continue
			}

			if build.IsNil(field.Type) && 0 == i {
				pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)
				if !f.Null {
					dst.Tab(1).Code("if nil == i." + fName)
					if build.GetBaseType(field.Type) == build.String {
						dst.Code(" || len(i.Get" + fName + "()) == 0")
					}
					dst.Code(" {\n")
					dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
					dst.Tab(1).Code("}\n")
				} else {
					dst.Tab(1).Code("if nil == i." + fName)
					if build.GetBaseType(field.Type) == build.String {
						dst.Code(" || len(i.Get" + fName + "()) == 0")
					}
					dst.Code(" {\n")
					dst.Tab(2).Code("return nil\n")
					dst.Tab(1).Code("}\n")
				}
			}
			if build.IsEnum(field.Type) {
				pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)
				dst.Tab(1).Code("if 0 == len(i.Get" + fName + "().ToName()) {\n")
				dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
				dst.Tab(1).Code("}\n")
			} else if build.IsMap(field.Type) {

			} else if build.IsArray(field.Type) {
				for _, item := range f.Len {
					pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)
					dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
					dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)
					dst.Tab(1).Code("if ").Code("!(len(i.").Code(fName).Code(") ").Code(b.getOperator(item.Op)).Code(" ").Code(item.Val).Code(") {\n")
					dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
					dst.Tab(1).Code("}\n")
				}
			} else {
				t := build.GetBaseType(field.Type)
				switch t {
				case build.Int8, build.Int16, build.Int32, build.Uint8, build.Uint16, build.Uint32, build.Float, build.Double, build.Uint64, build.Int64:
					for _, item := range f.Val {
						pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)

						dst.Tab(1).Code("if ").Code("!(i.Get").Code(fName).Code("() ").Code(b.getOperator(item.Op)).Code(" ").Code(item.Val).Code(") {\n")
						dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
				case build.Date:
					for _, item := range f.Val {
						pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)

						parse, err := time.Parse("2006-01-02T15:04:05Z", item.Val)
						if err != nil {
							return err
						}

						dst.Tab(1).Code("if ").Code("!(i.Get").Code(fName).Code("() ").Code(b.getOperator(item.Op)).Code(" ").Code(strconv.FormatInt(parse.UnixMilli(), 10)).Code(") {\n")
						dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
				case build.Decimal:
					for _, item := range f.Val {
						pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)
						dst.Import("github.com/shopspring/decimal", "", 0)

						dst.Tab(1).Code("if ")
						if build.OperatorGt == item.Op {
							dst.Code("!i.Get").Code(fName).Code("().").Code("GreaterThan")
						} else if build.OperatorLt == item.Op {
							dst.Code("!i.Get").Code(fName).Code("().").Code("LessThan")
						} else if build.OperatorGte == item.Op {
							dst.Code("!i.Get").Code(fName).Code("().").Code("GreaterThanOrEqual")
						} else if build.OperatorLte == item.Op {
							dst.Code("!i.Get").Code(fName).Code("().").Code("LessThanOrEqual")
						} else if build.OperatorEq == item.Op {
							dst.Code("!i.Get").Code(fName).Code("().").Code("Equal")
						} else if build.OperatorNeq == item.Op {
							dst.Code("i.Get").Code(fName).Code("().").Code("Equal")
						} else {
							//TODO
						}
						dst.Code("(decimal.NewFromFloat(").Code(item.Val).Code(")) {\n")
						dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}

				case build.String:
					for _, item := range f.Len {
						pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)
						dst.Import("unicode/utf8", "", 0)
						dst.Tab(1).Code("if !").Code("(utf8.RuneCountInString(i.Get").Code(fName).Code("()) ").Code(b.getOperator(item.Op)).Code(" ").Code(item.Val).Code(") {\n")
						dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
					for _, item := range f.Val {
						pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)
						if item.Op == build.OperatorMatch || item.Op == build.OperatorNotMatch {
							dst.Tab(1).Code("match, err ")
							if first {
								dst.Code(":")
							}
							dst.Import("regexp", "", 0)
							dst.Code("= regexp.MatchString(\"").Code(item.Val).Code("\", i.Get").Code(fName).Code("())\n")
							dst.Tab(1).Code("if err != nil {\n")
							dst.Tab(2).Code("return err\n")
							dst.Tab(1).Code("}\n")
							dst.Tab(1).Code("if false ").Code(b.getMatch(item.Op)).Code(" match {\n")
							dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
							dst.Tab(1).Code("}\n")
							first = false
						} else {
							dst.Tab(1).Code("if ").Code("!(i.Get").Code(fName).Code("() ").Code(b.getOperator(item.Op)).Code(" ").Code(item.Val).Code(") {\n")
							dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
							dst.Tab(1).Code("}\n")
						}
					}
				}
			}
		}

		dst.Tab(1).Code("return nil\n")
		dst.Code("}\n\n")
		return nil
	})
	if err != nil {
		return build.ErrorToFileError(err, b.fSet)
	}
	return nil
}

func (b *Builder) printVerifyDataCode(dst *build.Writer, data *ast.DataType) error {
	uName := build.StringToHumpName(data.Name.Name)
	lName := build.StringToFirstLower(data.Name.Name)
	b.getPackage(dst, data.Name)

	dst.Code("var ").Code(lName).Code("VerifyMaps = map[string]func(ctx context.Context, i *")
	dst.Code(uName).Code(" ,fields ...string) error {\n")
	err := build.EnumField(data, func(field *ast.Field, data *ast.DataType) error {
		_, ok := build.GetTag(field.Tags, "verify")
		if ok {
			dst.Tab(1).Code("\"").Code(field.Name.Name).Code("\": func(ctx context.Context, i *")
			dst.Code(uName).Code(" ,fields ...string) error {\n")
			dst.Tab(2).Code("return i.Verify").Code(build.StringToHumpName(field.Name.Name)).Code("(ctx, fields...)\n")
			dst.Tab(1).Code("},\n")
		}
		return nil
	})
	if err != nil {
		return err
	}
	dst.Code("}\n\n")

	dst.Tab(0).Code("func (i *").Code(uName).Code(") Verify(ctx context.Context, fields ...string) error {\n")
	dst.Tab(1).Code("if len(fields) == 0 {\n")
	dst.Tab(2).Code("for _, val := range ").Code(lName).Code("VerifyMaps {\n")
	dst.Tab(3).Code("if err := val(ctx, i); err != nil {\n")
	dst.Tab(4).Code("return err\n")
	dst.Tab(3).Code("}\n")
	dst.Tab(2).Code("}\n")
	dst.Tab(2).Code("return nil\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("for _, field := range fields {\n")
	dst.Tab(2).Code("if val, ok := ").Code(lName).Code("VerifyMaps[field]; ok {\n")
	dst.Tab(3).Code("if err := val(ctx, i); err != nil {\n")
	dst.Tab(4).Code("return err\n")
	dst.Tab(3).Code("}\n")
	dst.Tab(2).Code("}\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("return nil\n")
	dst.Tab(0).Code("}\n")
	return nil
}

func (b *Builder) getOperator(op build.Operator) string {
	if build.OperatorGt == op {
		return ">"
	} else if build.OperatorLt == op {
		return "<"
	} else if build.OperatorGte == op {
		return ">="
	} else if build.OperatorLte == op {
		return "<="
	} else if build.OperatorEq == op {
		return "=="
	} else if build.OperatorNeq == op {
		return "!="
	} else {
		//TODO 处理错误
	}
	return ""
}

func (b *Builder) getMatch(op build.Operator) string {
	if build.OperatorMatch == op {
		return "=="
	} else if build.OperatorNotMatch == op {
		return "!="
	} else {
		//TODO 处理错误
	}
	return ""
}
