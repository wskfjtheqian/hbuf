// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements NewPackage.

package ast

import (
	"fmt"
	"hbuf/pkg/scanner"
	"hbuf/pkg/token"
)

type pkgBuilder struct {
	fset   *token.FileSet
	errors scanner.ErrorList
}

func (p *pkgBuilder) error(pos token.Pos, msg string) {
	p.errors.Add(p.fset.Position(pos), msg)
}

func (p *pkgBuilder) errorf(pos token.Pos, format string, args ...interface{}) {
	p.error(pos, fmt.Sprintf(format, args...))
}

func (p *pkgBuilder) declare(scope, altScope *Scope, obj *Object) {
	alt := scope.Insert(obj)
	if alt == nil && altScope != nil {
		// see if there is a conflicting declaration in altScope
		alt = altScope.Lookup(obj.Name)
	}
	if alt != nil {
		prevDecl := ""
		if pos := alt.Pos(); pos.IsValid() {
			prevDecl = fmt.Sprintf("\n\tprevious declaration at %s", p.fset.Position(pos))
		}
		p.error(obj.Pos(), fmt.Sprintf("%s redeclared in this block%s", obj.Name, prevDecl))
	}
}

func resolve(scope *Scope, ident *Ident) bool {
	for ; scope != nil; scope = scope.Outer {
		if obj := scope.Lookup(ident.Name); obj != nil {
			ident.Obj = obj
			return true
		}
	}
	return false
}

type Importer func(imports map[string]*Object, path string) (pkg *Object, err error)
