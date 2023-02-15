package golang

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
)

func (b *Builder) printVerifyCode(dst *build.Writer, data *ast.DataType) error {
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

		dst.Code("func Verify" + dName + "_" + fName + "(val ")
		b.printType(dst, field.Type, false, false)
		dst.Code(") error {\n")

		for _, val := range verify.GetFormat() {
			f := build.GetFormat(val.Item.Tags)
			if nil == f {
				continue
			}

			if !f.IsNull && build.IsNil(field.Type) {
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
				pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)
				dst.Code("\tif nil == val {\n")
				dst.Code("\t\treturn &hbuf.Result{Code: int(" + pack + "), Msg: " + pack + ".ToName()}\n")
				dst.Code("\t}\n")

			}
			if 0 < len(f.Reg) {
				dst.Import("regexp", "")
				dst.Import("github.com/wskfjtheqian/hbuf_golang/pkg/hbuf", "")
				pack := b.getPackage(dst, val.Enum.Name) + build.StringToHumpName(val.Enum.Name.Name) + build.StringToHumpName(val.Item.Name.Name)

				if build.IsNil(field.Type) {
					dst.Code("\tif nil != val {\n")
					dst.Code("\t\tmatch, err := regexp.MatchString(\"" + f.Reg + "\", *val)\n")
					dst.Code("\t\tif err != nil {\n")
					dst.Code("\t\t\treturn err\n")
					dst.Code("\t\t}\n")
					dst.Code("\t\tif !match {\n")
					dst.Code("\t\t\treturn &hbuf.Result{Code: int(" + pack + "), Msg: " + pack + ".ToName()}\n")
					dst.Code("\t\t}\n")
					dst.Code("\t}\n")
				} else {
					dst.Code("\tmatch, err := regexp.MatchString(\"" + f.Reg + "\", val)\n")
					dst.Code("\tif err != nil {\n")
					dst.Code("\t\treturn err\n")
					dst.Code("\t}\n")
					dst.Code("\tif !match {\n")
					dst.Code("\t\treturn &hbuf.Result{Code: int(" + pack + "), Msg: " + pack + ".ToName()}\n")
					dst.Code("\t}\n")
				}
			}
		}
		dst.Code("\treturn nil\n")
		dst.Code("}\n\n")
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) printVerifyDataCode(dst *build.Writer, data *ast.DataType) error {
	dName := build.StringToHumpName(data.Name.Name)
	b.getPackage(dst, data.Name)

	dst.Code("func (i *" + dName + ") Verify() error {\n")

	isErr := true
	err := build.EnumField(data, func(field *ast.Field, data *ast.DataType) error {
		fName := build.StringToHumpName(field.Name.Name)
		_, ok := build.GetTag(field.Tags, "verify")
		if ok {
			if isErr {
				dst.Code("\tvar err error\n")
				isErr = false
			}
			dst.Code("\terr = Verify" + dName + "_" + fName + "(i." + fName + ")\n")
			dst.Code("\tif err != nil {\n")
			dst.Code("\t\treturn err\n")
			dst.Code("\t}\n")
		}
		return nil
	})
	if err != nil {
		return err
	}
	dst.Code("\treturn nil\n")
	dst.Code("}\n\n")
	return nil
}
