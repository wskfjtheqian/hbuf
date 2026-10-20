package wire

import (
	"hbuf/pkg/ast"
	"sort"
	"strconv"
	"strings"
)

// baseTypes HBUF 基础类型名集合(与编译前端一致;void/stream 为特殊基础类型)
var baseTypes = map[string]struct{}{
	"int8": {}, "int16": {}, "int32": {}, "int64": {},
	"uint8": {}, "uint16": {}, "uint32": {}, "uint64": {},
	"bool": {}, "float": {}, "double": {}, "decimal": {},
	"bytes": {}, "string": {}, "date": {},
	"void": {}, "stream": {},
}

// Convert 将整个包转换为中性描述,lang 为目标语言(决定引用类型的包名标注)
func Convert(pkg *ast.Package, lang string) []*File {
	files := make([]*File, 0, len(pkg.Files))
	for _, f := range pkg.Files {
		files = append(files, convertFile(pkg, f, lang))
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}

func convertFile(pkg *ast.Package, file *ast.File, lang string) *File {
	wf := &File{
		Path:     file.Path,
		Packages: map[string]string{},
		Imports:  make([]string, 0, len(file.Imports)),
	}
	for name, spec := range file.Packages {
		if nil != spec && nil != spec.Value {
			wf.Packages[name] = strings.Trim(spec.Value.Value, "\"")
		}
	}
	for _, imp := range file.Imports {
		if nil != imp && nil != imp.Path {
			wf.Imports = append(wf.Imports, strings.Trim(imp.Path.Value, "\""))
		}
	}
	sort.Strings(wf.Imports)
	for _, spec := range file.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		if typ := convertType(pkg, lang, ts.Type); typ != nil {
			wf.Types = append(wf.Types, typ)
		}
	}
	return wf
}

func convertType(pkg *ast.Package, lang string, expr ast.Expr) *Type {
	switch t := expr.(type) {
	case *ast.EnumType:
		et := &EnumType{TypeBase: convertTypeBase(t.Name.Name, t.Tags, nil, t.Doc)}
		for _, item := range t.Items {
			if nil == item {
				continue
			}
			et.Items = append(et.Items, &EnumItem{
				Name: item.Name.Name,
				Id:   parseInt(item.Id),
				Tags: convertTags(item.Tags),
				Doc:  doc(item.Doc),
			})
		}
		return &Type{Kind: "enum", Enum: et}
	case *ast.DataType:
		dt := &DataType{TypeBase: convertTypeBase(t.Name.Name, t.Tags, t.Extends, t.Doc)}
		for _, field := range t.Fields.List {
			if nil == field || nil == field.Name {
				continue
			}
			dt.Fields = append(dt.Fields, &Field{
				Name: field.Name.Name,
				Id:   parseInt(field.Id),
				Type: convertVarType(pkg, lang, field.Type),
				Tags: convertTags(field.Tags),
				Doc:  doc(field.Doc),
			})
		}
		return &Type{Kind: "data", Data: dt}
	case *ast.ServerType:
		st := &ServerType{TypeBase: convertTypeBase(t.Name.Name, t.Tags, t.Extends, t.Doc)}
		for _, method := range t.Methods {
			if nil == method || nil == method.Name {
				continue
			}
			m := &Method{
				Name: method.Name.Name,
				Id:   parseInt(method.Id),
				Tags: convertTags(method.Tags),
				Doc:  doc(method.Doc),
			}
			// Result 为 nil 或空类型表示广播(无返回值)
			if nil != method.Result && nil != method.Result.TypeExpr {
				m.Result = convertVarType(pkg, lang, method.Result)
			}
			if nil != method.Param && nil != method.Param.TypeExpr {
				m.Param = convertVarType(pkg, lang, method.Param)
			}
			if nil != method.ParamName {
				m.ParamName = method.ParamName.Name
			}
			st.Methods = append(st.Methods, m)
		}
		return &Type{Kind: "server", Server: st}
	}
	return nil
}

func convertTypeBase(name string, tags []*ast.Tag, extends []*ast.Extends, group *ast.CommentGroup) TypeBase {
	tb := TypeBase{
		Name: name,
		Tags: convertTags(tags),
		Doc:  doc(group),
	}
	for _, e := range extends {
		if nil == e || nil == e.Name {
			continue
		}
		tb.Extends = append(tb.Extends, &Extends{Name: e.Name.Name, Id: parseInt(e.Id)})
	}
	return tb
}

// convertVarType 把 ast 的 VarType/ArrayType/MapType/Ident 统一扁平化为 VarType
func convertVarType(pkg *ast.Package, lang string, expr ast.Expr) *VarType {
	switch t := expr.(type) {
	case *ast.VarType:
		vt := convertVarType(pkg, lang, t.TypeExpr)
		if nil != vt {
			vt.Empty = vt.Empty || t.Empty
		}
		return vt
	case *ast.ArrayType:
		vt := &VarType{Kind: "array", Empty: t.Empty}
		if nil != t.VType {
			vt.Element = convertVarType(pkg, lang, t.VType)
		}
		return vt
	case *ast.MapType:
		vt := &VarType{Kind: "map", Key: typeName(t.Key), Empty: t.Empty}
		if nil != t.VType {
			vt.Value = convertVarType(pkg, lang, t.VType)
		}
		return vt
	case *ast.Ident:
		if _, ok := baseTypes[t.Name]; ok {
			return &VarType{Kind: "base", Name: t.Name}
		}
		vt := &VarType{Kind: "ref", Name: t.Name, RefPack: refPack(pkg, lang, t.Name)}
		if nil != t.Obj && ast.Enum == t.Obj.Kind {
			vt.RefKind = "enum"
		} else if ast.Enum == objKind(pkg, t.Name) {
			vt.RefKind = "enum"
		} else {
			vt.RefKind = "data"
		}
		return vt
	}
	return &VarType{Kind: "base"}
}

// objKind 在包内查找类型声明的 ObjKind(跨文件引用时 ast 未解析引用)
func objKind(pkg *ast.Package, name string) ast.ObjKind {
	for _, f := range pkg.Files {
		if nil != f.Scope {
			if obj := f.Scope.Lookup(name); nil != obj {
				return obj.Kind
			}
		}
	}
	return ast.Bad
}

// typeName 提取 map key 的类型名
func typeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.VarType:
		return typeName(t.TypeExpr)
	}
	return ""
}

// refPack 查找引用类型所属文件在目标语言下的包名
func refPack(pkg *ast.Package, lang string, name string) string {
	for _, f := range pkg.Files {
		if nil != f.Scope && nil != f.Scope.Lookup(name) {
			if spec, ok := f.Packages[lang]; ok && nil != spec.Value {
				return strings.Trim(spec.Value.Value, "\"")
			}
			return ""
		}
	}
	return ""
}

func convertTags(tags []*ast.Tag) []*Tag {
	if nil == tags {
		return nil
	}
	ret := make([]*Tag, 0, len(tags))
	for _, tag := range tags {
		if nil == tag || nil == tag.Name {
			continue
		}
		t := &Tag{Name: tag.Name.Name}
		for _, kv := range tag.KV {
			if nil == kv || nil == kv.Name {
				continue
			}
			item := &KeyValue{Name: kv.Name.Name}
			for _, val := range kv.Values {
				if nil == val {
					continue
				}
				item.Values = append(item.Values, strings.Trim(val.Value, "\""))
			}
			t.KV = append(t.KV, item)
		}
		ret = append(ret, t)
	}
	return ret
}

func parseInt(lit *ast.BasicLit) int64 {
	if nil == lit {
		return 0
	}
	v, err := strconv.ParseInt(lit.Value, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func doc(group *ast.CommentGroup) string {
	if nil == group {
		return ""
	}
	return strings.TrimSpace(group.Text())
}
