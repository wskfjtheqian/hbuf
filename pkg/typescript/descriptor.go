package ts

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/build"
	"path/filepath"
)

// printDataDescriptor 生成 descriptor 单例：
// new SyncDescriptor 包裹（延迟构建，与 Go 生成器一致，支持循环引用），
// fields 按 ID 挂入 FieldDesc Map，extends 挂入 extends_ Map（键 = 扩展 ID）。
func (b *Builder) printDataDescriptor(dst *build.Writer, typ *ast.DataType) error {
	name := descName(typ.Name.Name)
	uName := className(typ.Name.Name)

	dst.Import("hbuf_ts", "type FieldDesc", "DataDescriptor")
	dst.Code("export const " + name + ": DataDescriptor =  new DataDescriptor(\n")
	dst.Tab(1).Code("() => new " + uName + "(),\n")
	dst.Tab(1).Code("new Map<number, FieldDesc>([\n")
	for _, field := range typ.Fields.List {
		dst.Tab(2).Code("[" + field.Id.Value + ", {name: \"" + field.Name.Name + "\", desc: ")
		b.printDescriptor(dst, field.Type, false)
		dst.Code(", get: (o: " + uName + ") => ")
		b.printFieldGet(dst, field.Type, field)
		dst.Code(", set: (o: " + uName + ", v) => ")
		b.printFieldSet(dst, field.Type, field)
		dst.Code("}],\n")
	}
	dst.Tab(1).Code("]),\n")
	if 0 < len(typ.Extends) {
		dst.Import("hbuf_ts", "RefDescriptor")
		dst.Tab(1).Code("new Map<number, FieldDesc>([\n")
		for _, extend := range typ.Extends {
			dst.Tab(2).Code("[" + extend.Id.Value + ", {desc: new RefDescriptor(():Descriptor => ")
			b.getPackageDesc(dst, extend.Name)
			// 类为平铺继承（父类字段直接平铺在子类上），get 返回对象本身，decode 就地填充父类字段
			dst.Code(descName(extend.Name.Name) + "), get: (o: " + uName + ") => o, set: (o: " + uName + ", v) => {}}],\n")
		}
		dst.Tab(1).Code("]),\n")
	}
	dst.Code(")\n\n")
	return nil
}

// printDescriptor 生成字段描述符表达式（对应 Go 版 printDescriptor，无 unsafe.Offsetof）
func (b *Builder) printDescriptor(dst *build.Writer, expr ast.Expr, isNull bool) {
	isPtr := "false"
	if isNull {
		isPtr = "true"
	}
	switch expr.(type) {
	case *ast.Ident:
		t := expr.(*ast.Ident)
		if nil != t.Obj {
			if ast.Enum == t.Obj.Kind {
				dst.Import("hbuf_ts", "Int32Descriptor")
				dst.Code("new Int32Descriptor(" + isPtr + ")")
			} else {
				dst.Import("hbuf_ts", "RefDescriptor")
				b.getPackageDesc(dst, t)
				dst.Code("new RefDescriptor(()=>" + descName(t.Name) + ", " + isPtr + ")")
			}
		} else {
			switch build.BaseType(t.Name) {
			case build.Int8:
				dst.Import("hbuf_ts", "Int8Descriptor")
				dst.Code("new Int8Descriptor(" + isPtr + ")")
			case build.Int16:
				dst.Import("hbuf_ts", "Int16Descriptor")
				dst.Code("new Int16Descriptor(" + isPtr + ")")
			case build.Int32:
				dst.Import("hbuf_ts", "Int32Descriptor")
				dst.Code("new Int32Descriptor(" + isPtr + ")")
			case build.Int64:
				dst.Import("hbuf_ts", "Int64Descriptor")
				dst.Code("new Int64Descriptor(" + isPtr + ")")
			case build.Uint8:
				dst.Import("hbuf_ts", "Uint8Descriptor")
				dst.Code("new Uint8Descriptor(" + isPtr + ")")
			case build.Uint16:
				dst.Import("hbuf_ts", "Uint16Descriptor")
				dst.Code("new Uint16Descriptor(" + isPtr + ")")
			case build.Uint32:
				dst.Import("hbuf_ts", "Uint32Descriptor")
				dst.Code("new Uint32Descriptor(" + isPtr + ")")
			case build.Uint64:
				dst.Import("hbuf_ts", "Uint64Descriptor")
				dst.Code("new Uint64Descriptor(" + isPtr + ")")
			case build.Float:
				dst.Import("hbuf_ts", "FloatDescriptor")
				dst.Code("new FloatDescriptor(" + isPtr + ")")
			case build.Double:
				dst.Import("hbuf_ts", "DoubleDescriptor")
				dst.Code("new DoubleDescriptor(" + isPtr + ")")
			case build.Bool:
				dst.Import("hbuf_ts", "BoolDescriptor")
				dst.Code("new BoolDescriptor(" + isPtr + ")")
			case build.String:
				dst.Import("hbuf_ts", "StringDescriptor")
				dst.Code("new StringDescriptor(" + isPtr + ")")
			case build.Decimal:
				dst.Import("hbuf_ts", "DecimalDescriptor")
				dst.Code("new DecimalDescriptor(" + isPtr + ")")
			case build.Date:
				dst.Import("hbuf_ts", "TimeDescriptor")
				dst.Code("new TimeDescriptor(" + isPtr + ")")
			case build.Bytes:
				dst.Import("hbuf_ts", "BytesDescriptor")
				dst.Code("new BytesDescriptor(" + isPtr + ")")
			}
		}
	case *ast.ArrayType:
		ar := expr.(*ast.ArrayType)
		dst.Import("hbuf_ts", "ListDescriptor")
		dst.Code("new ListDescriptor(")
		b.printDescriptor(dst, ar.VType, false)
		dst.Code(", " + isPtr + ")")
	case *ast.MapType:
		ma := expr.(*ast.MapType)
		dst.Import("hbuf_ts", "MapDescriptor")
		dst.Code("new MapDescriptor(")
		b.printDescriptor(dst, ma.Key, false)
		dst.Code(", ")
		b.printDescriptor(dst, ma.VType, false)
		dst.Code(", " + isPtr + ")")
	case *ast.VarType:
		t := expr.(*ast.VarType)
		b.printDescriptor(dst, t.Type(), isNull || t.Empty)
	}
}

// printFieldGet 生成字段读取闭包（枚举需取 .value 适配数值描述符；非空枚举不判空）
func (b *Builder) printFieldGet(dst *build.Writer, expr ast.Expr, field *ast.Field) {
	lName := build.StringToFirstLower(field.Name.Name)
	if build.IsEnum(expr) {
		if build.IsNil(expr) {
			dst.Code("o." + lName + " == null ? null : o." + lName + ".value")
		} else {
			dst.Code("o." + lName + ".value")
		}
	} else {
		dst.Code("o." + lName)
	}
}

// printFieldSet 生成字段写入闭包（枚举需 valueOf 还原对象；非空枚举不产生 null 分支）
func (b *Builder) printFieldSet(dst *build.Writer, expr ast.Expr, field *ast.Field) {
	lName := build.StringToFirstLower(field.Name.Name)
	if build.IsEnum(expr) {
		b.getEnumPackage(dst, expr)
		if build.IsNil(expr) {
			dst.Code("o." + lName + " = v == null ? null : " + b.enumTypeName(expr) + ".valueOf(v)")
		} else {
			dst.Code("o." + lName + " = " + b.enumTypeName(expr) + ".valueOf(v)")
		}
	} else {
		dst.Code("o." + lName + " = v")
	}
}

// getEnumPackage 注册枚举类 import 并返回枚举类名
func (b *Builder) getEnumPackage(dst *build.Writer, expr ast.Expr) {
	var ident *ast.Ident
	switch expr.(type) {
	case *ast.Ident:
		ident = expr.(*ast.Ident)
	case *ast.VarType:
		ident = expr.(*ast.VarType).Type().(*ast.Ident)
	}
	if nil != ident && nil != ident.Obj {
		b.getPackage(dst, ident, "", false, false, "", "", nil)
	}
}

// enumTypeName 解出枚举类名（枚举类名不带 Ts 后缀）
func (b *Builder) enumTypeName(expr ast.Expr) string {
	var ident *ast.Ident
	switch expr.(type) {
	case *ast.Ident:
		ident = expr.(*ast.Ident)
	case *ast.VarType:
		ident = expr.(*ast.VarType).Type().(*ast.Ident)
	}
	if nil != ident {
		return ident.Name
	}
	return ""
}

// getPackageDesc 跨文件引用 data 的 desc 单例 import
func (b *Builder) getPackageDesc(dst *build.Writer, expr ast.Expr) {
	file := (expr.(*ast.Ident)).Obj.Data
	switch file.(type) {
	case *ast.File:
		break
	default:
		return
	}

	_, name := filepath.Split(file.(*ast.File).Path)
	name = name[:len(name)-len(".hbuf")]
	name = name + ".data"
	dst.Import("./"+name, descName(expr.(*ast.Ident).Name))
}
