package format

import (
	"fmt"
	"hbuf/pkg/ast"
	"hbuf/pkg/parser"
	"hbuf/pkg/token"
	"os"
)

func ExampleFormat_format() {
	fset := token.NewFileSet() // positions are relative to fset
	src := []byte("" +
		"package \"parser\" \n" +
		"//引用parser.go \n" +
		"import \"/home/yttx_heqian/develop/go/hbuf/pkg/parser/parser.go\" \n" +
		"//引用11.go \n" +
		"import \"/home/yttx_heqian/develop/go/hbuf/pkg/parser/11.go\" \n" +
		"data NAME : Na,Nb { \n" +
		"  String Type = 16 `json\"name\"` //姓名\n" +
		"  String[] Info = 0 \n" +
		"  String<int> other = 0 `json\"map\"` //姓名\n" +
		"} \n" +
		"\n " +
		"data Type : Type{ \n" +
		" int age = 15 `pr:id,json\"age\"` \n" +
		"} \n" +
		"\n" +
		"server GetName{ \n" +
		"   String name(int Id) \n" +
		"} \n" +
		"\n" +
		"server GetAge : GetName{ \n" +
		"   Int age(int Id = 1,String key = 2) \n" +
		"} \n" +
		"enum Type{ \n" +
		"   New" +
		"   Old" +
		"} \n" +
		"")

	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		fmt.Println(err)
		return
	}

	ast.SortImports(fset, f)
	Node(os.Stdout, f)
	// output:
}
