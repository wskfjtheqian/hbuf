// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements scopes and the objects they contain.

package ast

import (
	"bytes"
	"fmt"
	"hbuf/pkg/token"
)

type Scope struct {
	Outer   *Scope
	Objects map[string]*Object
}

func NewScope(outer *Scope) *Scope {
	const n = 4 // initial scope capacity
	return &Scope{outer, make(map[string]*Object, n)}
}

func (s *Scope) Lookup(name string) *Object {
	return s.Objects[name]
}

func (s *Scope) Insert(obj *Object) (alt *Object) {
	if alt = s.Objects[obj.Name]; alt == nil {
		s.Objects[obj.Name] = obj
	}
	return
}

func (s *Scope) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "scope %p {", s)
	if s != nil && len(s.Objects) > 0 {
		fmt.Fprintln(&buf)
		for _, obj := range s.Objects {
			fmt.Fprintf(&buf, "\t%s %s\n", obj.Kind, obj.Name)
		}
	}
	fmt.Fprintf(&buf, "}\n")
	return buf.String()
}

// ----------------------------------------------------------------------------
// Objects

// An Object describes a named language entity such as a package,
// constant, type, variable, function (incl. methods), or label.
//
// The Data fields contains object-specific data:
//
//	Kind    Data type         Data value
//	Pkg     *Scope            package scope
//	Con     int               iota for the respective declaration
//
type Object struct {
	Kind ObjKind
	Name string      // declared name
	Decl interface{} // corresponding Field, XxxSpec, FuncDecl, LabeledStmt, AssignStmt, Scope; or nil
	Data interface{} // object-specific data; or nil
	Type interface{} // placeholder for type information; may be nil
}

// NewObj creates a new object of a given kind and name.
func NewObj(kind ObjKind, name string) *Object {
	return &Object{Kind: kind, Name: name}
}

// Pos computes the source position of the declaration of an object name.
// The result may be an invalid position if it cannot be computed
// (obj.Decl may be nil or not correct).
func (obj *Object) Pos() token.Pos {
	//name := obj.Name
	//switch d := obj.Decl.(type) {
	//case *Field:
	//	//for _, n := range d.Names {
	//	//	if n.Name == name {
	//	//		return n.Pos()
	//	//	}
	//	//}
	//case *ImportSpec:
	//case *Scope:
	//	// predeclared object - nothing to do for now
	//}
	return token.NoPos
}

// ObjKind describes what an object represents.
type ObjKind int

// The list of possible Object kinds.
const (
	Bad ObjKind = iota // for error handling
	Pkg                // package
	Data
	Server
	Method // function or method
	Var
	Enum
	Int8
	Int16
	Int32
	Int64
	Uint8
	Uint16
	Uint32
	Uint64
	Bool
	Float
	Double
	String
)

var objKindStrings = [...]string{
	Bad:    "bad",
	Pkg:    "package",
	Data:   "data",
	Server: "server",
	Method: "method",
	Var:    "var",
	Enum:   "enum",
	Int8:   "int8",
	Int16:  "int16",
	Int32:  "int32",
	Int64:  "int64",
	Uint8:  "uint8",
	Uint16: "uint16",
	Uint32: "uint32",
	Uint64: "uint64",
	Bool:   "bool",
	Float:  "float",
	Double: "double",
	String: "string",
}

func (kind ObjKind) String() string { return objKindStrings[kind] }
