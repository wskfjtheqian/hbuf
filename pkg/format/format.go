package format

import (
	"fmt"
	"hbuf/pkg/ast"
	"io"
	"sort"
	"strings"
)

// Node 将 AST 节点序列化为 hbuf 源码,当前仅支持 *ast.File。
func Node(dst io.Writer, node interface{}) error {
	file, ok := node.(*ast.File)
	if !ok {
		return fmt.Errorf("unsupported node type %T", node)
	}
	return File(dst, file)
}

// File 将整个文件序列化为 hbuf 源码。
func File(dst io.Writer, file *ast.File) error {
	if nil == file {
		return fmt.Errorf("file is nil")
	}
	// Packages 为 map,按语言名排序保证输出确定。
	names := make([]string, 0, len(file.Packages))
	for name := range file.Packages {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := printPackage(dst, file.Packages[name]); err != nil {
			return err
		}
	}
	if len(names) > 0 {
		io.WriteString(dst, "\n")
	}

	for _, spec := range file.Imports {
		if err := printImport(dst, spec); err != nil {
			return err
		}
	}
	if len(file.Imports) > 0 {
		io.WriteString(dst, "\n")
	}

	for _, spec := range file.Specs {
		if typ, ok := spec.(*ast.TypeSpec); ok {
			if err := printTypeSpec(dst, typ); err != nil {
				return err
			}
			io.WriteString(dst, "\n")
		}
	}
	return nil
}

func printPackage(dst io.Writer, spec *ast.PackageSpec) error {
	if nil == spec || nil == spec.Name || nil == spec.Value {
		return nil
	}
	_, err := fmt.Fprintf(dst, "package %s = %s;\n", spec.Name.Name, spec.Value.Value)
	return err
}

func printImport(dst io.Writer, spec *ast.ImportSpec) error {
	if nil == spec || nil == spec.Path {
		return nil
	}
	_, err := fmt.Fprintf(dst, "import %s;\n", spec.Path.Value)
	return err
}

func printTypeSpec(dst io.Writer, spec *ast.TypeSpec) error {
	if spec.Doc != nil {
		printDoc(dst, spec.Doc)
	}
	switch typ := spec.Type.(type) {
	case *ast.DataType:
		return printData(dst, typ)
	case *ast.ServerType:
		return printServer(dst, typ)
	case *ast.EnumType:
		return printEnum(dst, typ)
	default:
		return fmt.Errorf("unsupported type %T", spec.Type)
	}
}

func printData(dst io.Writer, typ *ast.DataType) error {
	printTags(dst, typ.Tags)
	io.WriteString(dst, "data "+typ.Name.Name)
	printExtends(dst, typ.Extends)
	io.WriteString(dst, " {\n")
	for _, field := range typ.Fields.List {
		if field.Doc != nil {
			printDoc(dst, field.Doc)
		}
		io.WriteString(dst, "    ")
		printType(dst, field.Type)
		io.WriteString(dst, " "+field.Name.Name+" = "+field.Id.Value)
		if nil != field.Tag {
			io.WriteString(dst, " "+field.Tag.Value)
		}
		io.WriteString(dst, ";")
		if field.Comment != nil {
			io.WriteString(dst, " // "+field.Comment.Text())
		}
		io.WriteString(dst, "\n")
	}
	io.WriteString(dst, "};")
	if typ.Comment != nil {
		io.WriteString(dst, " // "+typ.Comment.Text())
	}
	io.WriteString(dst, "\n")
	return nil
}

func printServer(dst io.Writer, typ *ast.ServerType) error {
	printTags(dst, typ.Tags)
	io.WriteString(dst, "server "+typ.Name.Name)
	printExtends(dst, typ.Extends)
	io.WriteString(dst, " {\n")
	for _, method := range typ.Methods {
		if method.Doc != nil {
			printDoc(dst, method.Doc)
		}
		io.WriteString(dst, "    ")
		printType(dst, method.Result)
		io.WriteString(dst, " "+method.Name.Name+"(")
		printType(dst, method.Param)
		io.WriteString(dst, " "+method.ParamName.Name)
		io.WriteString(dst, ") = "+method.Id.Value+";\n")
	}
	io.WriteString(dst, "};")
	if typ.Comment != nil {
		io.WriteString(dst, " // "+typ.Comment.Text())
	}
	io.WriteString(dst, "\n")
	return nil
}

func printEnum(dst io.Writer, typ *ast.EnumType) error {
	printTags(dst, typ.Tags)
	io.WriteString(dst, "enum "+typ.Name.Name+" {\n")
	for _, item := range typ.Items {
		if item.Doc != nil {
			printDoc(dst, item.Doc)
		}
		io.WriteString(dst, "    "+item.Name.Name+" = "+item.Id.Value+";")
		if item.Comment != nil {
			io.WriteString(dst, " // "+item.Comment.Text())
		}
		io.WriteString(dst, "\n")
	}
	io.WriteString(dst, "};")
	if typ.Comment != nil {
		io.WriteString(dst, " // "+typ.Comment.Text())
	}
	io.WriteString(dst, "\n")
	return nil
}

func printExtends(dst io.Writer, extends []*ast.Extends) {
	if len(extends) == 0 {
		return
	}
	io.WriteString(dst, " : ")
	for i, e := range extends {
		if 0 != i {
			io.WriteString(dst, ", ")
		}
		io.WriteString(dst, e.Name.Name+" = "+e.Id.Value)
	}
}

func printType(dst io.Writer, expr ast.Expr) {
	switch t := expr.(type) {
	case *ast.VarType:
		io.WriteString(dst, identName(t.TypeExpr))
		if t.Empty {
			io.WriteString(dst, "?")
		}
	case *ast.ArrayType:
		printType(dst, t.VType)
		io.WriteString(dst, "[]")
		if t.Empty {
			io.WriteString(dst, "?")
		}
	case *ast.MapType:
		printType(dst, t.VType)
		io.WriteString(dst, "<")
		printType(dst, t.Key)
		io.WriteString(dst, ">")
		if t.Empty {
			io.WriteString(dst, "?")
		}
	default:
		io.WriteString(dst, identName(expr))
	}
}

func printTags(dst io.Writer, tags []*ast.Tag) {
	for _, tag := range tags {
		io.WriteString(dst, "["+tag.Name.Name+":")
		for _, kv := range tag.KV {
			io.WriteString(dst, " "+kv.Name.Name+"=")
			for j, v := range kv.Values {
				if 0 != j {
					io.WriteString(dst, ",")
				}
				io.WriteString(dst, v.Value)
			}
		}
		io.WriteString(dst, "];\n")
	}
}

func printDoc(dst io.Writer, doc *ast.CommentGroup) {
	for _, line := range strings.Split(doc.Text(), "\n") {
		io.WriteString(dst, "// "+line+"\n")
	}
}

func identName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok && nil != ident {
		return ident.Name
	}
	return ""
}
