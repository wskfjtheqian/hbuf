package golang

import (
	"fmt"
	"hbuf/pkg/ast"
	"hbuf/pkg/parser"
	"hbuf/pkg/token"
	"os"
)

func ExampleGolang_out() {
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
		" int Age = 15 `pr:id,json\"Age\"` \n" +
		"} \n" +
		"\n" +
		"server GetName{ \n" +
		"   String name(int Id) \n" +
		"} \n" +
		"\n" +
		"server GetAge : GetName{ \n" +
		"   Int Age(int Id = 1,String key = 2) \n" +
		"} \n")

	f, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		fmt.Println(err)
		return
	}

	ast.SortImports(fset, f)
	Node(os.Stdout, f)
	// output:
}
