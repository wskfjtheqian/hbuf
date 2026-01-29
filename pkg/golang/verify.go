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

			pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)

			if build.IsNil(field.Type) && 0 == i {
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
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)
				dst.Tab(1).Code("if 0 == len(i.Get" + fName + "().ToName()) {\n")
				dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
				dst.Tab(1).Code("}\n")
			} else if build.IsMap(field.Type) {

			} else if build.IsArray(field.Type) {

			} else {
				t := build.GetBaseType(field.Type)
				switch t {
				case build.Int8, build.Int16, build.Int32, build.Uint8, build.Uint16, build.Uint32, build.Float, build.Double:
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)
						dst.Tab(1).Code("if ")
						if 0 < len(f.Min) {
							dst.Code(f.Min + " > i.Get" + fName + "() ")
						}
						if 0 < len(f.Max) {
							if 0 < len(f.Min) {
								dst.Code("|| ")
							}
							dst.Code(f.Max + " < i.Get" + fName + "() ")
						}
						dst.Code("{\n")
						dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
				case build.Uint64, build.Int64:
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)
						dst.Tab(1).Code("if ")
						if 0 < len(f.Min) {
							dst.Code(f.Min + " > i.Get" + fName + "().Val ")
						}
						if 0 < len(f.Max) {
							if 0 < len(f.Min) {
								dst.Code("|| ")
							}
							dst.Code(f.Max + " < i.Get" + fName + "().Val ")
						}
						dst.Code("{\n")
						dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
				case build.Date:
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)
						dst.Tab(1).Code("if ")
						if 0 < len(f.Min) {
							parse, err := time.Parse("2006-01-02T15:04:05Z", f.Min)
							if err != nil {
								return err
							}
							dst.Code(strconv.FormatInt(parse.UnixMilli(), 10) + " > i.Get" + fName + "().UnixMilli() ")
						}
						if 0 < len(f.Max) {
							if 0 < len(f.Min) {
								dst.Code("|| ")
							}
							parse, err := time.Parse("2006-01-02T15:04:05Z", f.Max)
							if err != nil {
								return err
							}
							dst.Code(strconv.FormatInt(parse.UnixMilli(), 10) + " < i.Get" + fName + "().UnixMilli() ")
						}
						dst.Code("{ //" + f.Min + "--" + f.Max + "\n")
						dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
				case build.Decimal:
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
						dst.Tab(1).Code("if ")
						if 0 < len(f.Min) {
							dst.Import("github.com/shopspring/decimal", "", 0)
							dst.Code("decimal.NewFromFloat(" + f.Min + ").GreaterThan(i.Get" + fName + "()) ")
						}
						if 0 < len(f.Max) {
							if 0 < len(f.Min) {
								dst.Code("|| ")
							}
							dst.Import("github.com/shopspring/decimal", "", 0)
							dst.Code("decimal.NewFromFloat(" + f.Min + ").LessThan(i.Get" + fName + "()) ")
						}
						dst.Code("{\n")
						dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
				case build.String:
					dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "", 0)
					dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hrpc", "", 0)
					pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Tab(1).Code("if ")
						if 0 < len(f.Min) {
							dst.Import("unicode/utf8", "", 0)
							dst.Code(f.Min + " > utf8.RuneCountInString(i.Get" + fName + "()) ")
						}
						if 0 < len(f.Max) {
							if 0 < len(f.Min) {
								dst.Code("|| ")
							}
							dst.Import("unicode/utf8", "", 0)
							dst.Code(f.Max + " < utf8.RuneCountInString(i.Get" + fName + "()) ")
						}
						dst.Code("{\n")
						dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
					if 0 < len(f.Reg) {
						dst.Tab(1).Code("match, err ")
						if first {
							dst.Code(":")
						}
						dst.Import("regexp", "", 0)
						dst.Code("= regexp.MatchString(\"" + f.Reg + "\", i.Get" + fName + "())\n")
						dst.Tab(1).Code("if err != nil {\n")
						dst.Tab(2).Code("return err\n")
						dst.Tab(1).Code("}\n")
						dst.Tab(1).Code("if !match {\n")
						dst.Tab(2).Code("return hrpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
						first = false
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
	dst.Code(uName).Code(" ,fields ...string) error {")
	err := build.EnumField(data, func(field *ast.Field, data *ast.DataType) error {
		_, ok := build.GetTag(field.Tags, "verify")
		if ok {
			dst.Tab(1).Code("\n\"").Code(field.Name.Name).Code("\": func(ctx context.Context, i *")
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
	dst.Tab(3).Code("if err := val(ctx, i); err == nil {\n")
	dst.Tab(4).Code("return err\n")
	dst.Tab(3).Code("}\n")
	dst.Tab(2).Code("}\n")
	dst.Tab(2).Code("return nil\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("for _, field := range fields {\n")
	dst.Tab(2).Code("if val, ok := ").Code(lName).Code("VerifyMaps[field]; ok {\n")
	dst.Tab(3).Code("if err := val(ctx, i); err == nil {\n")
	dst.Tab(4).Code("return err\n")
	dst.Tab(3).Code("}\n")
	dst.Tab(2).Code("}\n")
	dst.Tab(1).Code("}\n")
	dst.Tab(1).Code("return nil\n")
	dst.Tab(0).Code("}\n")
	return nil
}
