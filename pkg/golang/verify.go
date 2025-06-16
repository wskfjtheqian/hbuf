package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"strconv"
	"time"
)

func (b *Builder) printVerifyCode(dst *build.Writer, data *ast.DataType) error {
	dst.Import("context", "")
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

		dst.Code("func (i *" + dName + ") Verify" + fName + "(ctx context.Context) error {\n")

		first := true
		for i, val := range verify.GetFormat() {
			f := build.GetFormat(val.Item.Tags)
			if nil == f {
				continue
			}

			pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)

			if build.IsNil(field.Type) && 0 == i {
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/rpc", "")
				if !f.Null {
					dst.Tab(1).Code("if nil == i." + fName)
					if build.GetBaseType(field.Type) == build.String {
						dst.Code(" || len(i.Get" + fName + "()) == 0")
					}
					dst.Code(" {\n")
					dst.Tab(2).Code("return rpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
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
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/rpc", "")
				dst.Tab(1).Code("if 0 == len(i.Get" + fName + "().ToName()) {\n")
				dst.Tab(2).Code("return rpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
				dst.Tab(1).Code("}\n")
			} else if build.IsMap(field.Type) {

			} else if build.IsArray(field.Type) {

			} else {
				t := build.GetBaseType(field.Type)
				switch t {
				case build.Int8, build.Int16, build.Int32, build.Uint8, build.Uint16, build.Uint32, build.Float, build.Double:
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/rpc", "")
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
						dst.Tab(2).Code("return rpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
				case build.Uint64, build.Int64:
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/rpc", "")
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
						dst.Tab(2).Code("return rpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
				case build.Date:
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/rpc", "")
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
						dst.Tab(2).Code("return rpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
				case build.Decimal:
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
						dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/rpc", "")
						dst.Tab(1).Code("if ")
						if 0 < len(f.Min) {
							dst.Code("decimal.NewFromFloat(" + f.Min + ").GreaterThan(i.Get" + fName + "()) ")
						}
						if 0 < len(f.Max) {
							if 0 < len(f.Min) {
								dst.Code("|| ")
							}
							dst.Code("decimal.NewFromFloat(" + f.Min + ").LessThan(i.Get" + fName + "()) ")
						}
						dst.Code("{\n")
						dst.Tab(2).Code("return rpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
				case build.String:
					dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
					dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/rpc", "")
					pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)
					if 0 < len(f.Min) || 0 < len(f.Max) {
						dst.Tab(1).Code("if ")
						if 0 < len(f.Min) {
							dst.Import("unicode/utf8", "")
							dst.Code(f.Min + " > utf8.RuneCountInString(i.Get" + fName + "()) ")
						}
						if 0 < len(f.Max) {
							if 0 < len(f.Min) {
								dst.Code("|| ")
							}
							dst.Import("unicode/utf8", "")
							dst.Code(f.Max + " < utf8.RuneCountInString(i.Get" + fName + "()) ")
						}
						dst.Code("{\n")
						dst.Tab(2).Code("return rpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
						dst.Tab(1).Code("}\n")
					}
					if 0 < len(f.Reg) {
						dst.Tab(1).Code("match, err ")
						if first {
							dst.Code(":")
						}
						dst.Import("regexp", "")
						dst.Code("= regexp.MatchString(\"" + f.Reg + "\", i.Get" + fName + "())\n")
						dst.Tab(1).Code("if err != nil {\n")
						dst.Tab(2).Code("return err\n")
						dst.Tab(1).Code("}\n")
						dst.Tab(1).Code("if !match {\n")
						dst.Tab(2).Code("return rpc.NewResult[hbuf.Data](int32(").Code(pack).Code("), ").Code(pack).Code(".ToName(), nil)\n")
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
	dName := build.StringToHumpName(data.Name.Name)
	b.getPackage(dst, data.Name)

	dst.Code("func (i *" + dName + ") Verify(ctx context.Context) error {\n")

	isErr := true
	err := build.EnumField(data, func(field *ast.Field, data *ast.DataType) error {
		fName := build.StringToHumpName(field.Name.Name)
		_, ok := build.GetTag(field.Tags, "verify")
		if ok {
			if isErr {
				dst.Tab(1).Code("var err error\n")
				isErr = false
			}
			dst.Tab(1).Code("err = i.Verify" + fName + "(ctx)\n")
			dst.Tab(1).Code("if err != nil {\n")
			dst.Tab(2).Code("return err\n")
			dst.Tab(1).Code("}\n")
		}
		return nil
	})
	if err != nil {
		return err
	}
	dst.Tab(1).Code("return nil\n")
	dst.Code("}\n\n")
	return nil
}
