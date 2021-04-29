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
		"  String Name = 16 `json\"name\"` \n" +
		"  String[12] Info = 0 \n" +
		"  String<int> other = 0 \n" +
		"} \n" +
		"\n " +
		"data Name : Name{ \n" +
		" int age = 15 `pr:id,json\"age\"` \n" +
		"} \n" +
		"\n" +
		"server GetName{ \n" +
		"   String name(int Id) \n" +
		"} \n" +
		"\n" +
		"server GetAge : GetName{ \n" +
		"   Int age(int Id) \n" +
		"} \n")

	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("package" + f.Package.Path.Value)
	for _, s := range f.Imports {
		fmt.Println("import" + s.Path.Value)
	}

	for _, s := range f.Comments {
		fmt.Println(s.Text())
	}

	// output:
}
