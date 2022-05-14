// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser_test

import (
	"fmt"
	"hbuf/pkg/parser"
	"hbuf/pkg/token"
)

func ExampleParseFile() {
	fset := token.NewFileSet() // positions are relative to fset
	src := []byte("" +
		"package \"parser\" \n" +
		"//引用parser.go \n" +
		"import \"/home/yttx_heqian/develop/go/hbuf/pkg/parser/parser.go\" \n" +
		"//引用11.go \n" +
		"import \"/home/yttx_heqian/develop/go/hbuf/pkg/parser/11.go\" \n" +
		"data NAME : Na,Nb { \n" +
		"  String? Type = 16  `json\"name\"` //姓名\n" +
		"  String[]? Info = 0 \n" +
		"  String<int?>? other = 0 \n" +
		"}")

	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("package" + f.Package.Value.Value)
	for _, s := range f.Imports {
		fmt.Println("import" + s.Path.Value)
	}

	for _, s := range f.Comments {
		fmt.Println(s.Text())
	}

	// output:
}
