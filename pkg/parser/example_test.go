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
		"package go=\"parser\" \n" +
		"[dbName:key=\"ry\";key=\"ry\"]" +
		"data NAME = 0 { \n" +
		"}",
	)

	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, s := range f.Imports {
		fmt.Println("import" + s.Path.Value)
	}

	for _, s := range f.Comments {
		fmt.Println(s.Text())
	}

	// output:
}
