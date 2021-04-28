// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scanner

import (
	"fmt"
	"hbuf/pkg/token"
)

func ExampleScanner_Scan() {
	// src is the input that we want to tokenize.
	src := []byte("" +
		"package \"parser\" \n" +
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

	// Initialize the
	var s Scanner                                   // positions are relative to fset
	fset := token.NewFileSet()                      // positions are relative to fset
	file := fset.AddFile("", fset.Base(), len(src)) // register input "file"
	s.Init(file, src, nil /* no error handler */, ScanComments)

	// Repeated calls to Scan yield the token sequence found in the input.
	for {
		pot, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		fmt.Printf("%s\t%s\t%q\n", fset.Position(pot), tok, lit)
	}

	// output:
}
