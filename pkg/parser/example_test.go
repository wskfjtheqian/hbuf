// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser_test

import (
	"fmt"
	"go/parser"
	"go/token"
)

func ExampleParseFile() {
	fset := token.NewFileSet() // positions are relative to fset
	src := []byte("" +
		"package parser \n" +
		"import \"/home/yttx_heqian/develop/go/hbuf/pkg/parser/parser.go\" \n" +
		"data NAME{ \n" +
		"  String Name = 16 `json\"name\"` \n" +
		"  String[] Info = 0 \n" +
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
	f, err := parser.ParseFile(fset, "", src, parser.ImportsOnly)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print the imports from the file's AST.
	for _, s := range f.Imports {
		fmt.Println(s.Path.Value)
	}

	// output:
}
