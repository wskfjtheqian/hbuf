package format

import (
	"fmt"
	"hbuf/pkg/ast"
	"hbuf/pkg/parser"
	"hbuf/pkg/token"
	"os"
)

func ExampleNode() {
	src := `package go = "admin";
import "parser.go";
data NAME : Na = 1 {
    string type = 16 "json:\"name\"";
    string[] info = 2;
    string<int> other = 3 "json:\"map\"";
};
data Type : NAME = 2 {
    int age = 15 "pr:id,json:\"age\"";
};
server GetName {
    string name(string id) = 1;
};
server GetAge : GetName = 2 {
    int age(string id) = 1;
};
enum Type {
    new = 0;
    old = 1;
};
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "base.hbuf", []byte(src), parser.AllErrors)
	if err != nil {
		fmt.Println(err)
		return
	}
	ast.SortImports(fset, f)
	if err := Node(os.Stdout, f); err != nil {
		fmt.Println(err)
	}
}
