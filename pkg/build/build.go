package build

import (
	"hbuf/pkg/ast"
	"hbuf/pkg/dart"
	"hbuf/pkg/parser"
	"hbuf/pkg/token"
	"path/filepath"
	"regexp"
	"strings"
)
import "hbuf/pkg/golang"

var buildInits = map[string]func(){
	"dart": dart.Init,
	"go":   golang.Init,
}

func CheckType(typ string) bool {
	_, ok := buildInits[typ]
	return ok
}

const (
	Int8   string = "int8"
	Int16  string = "int16"
	Int32  string = "int32"
	Int64  string = "int64"
	Uint8  string = "uint8"
	Uint16 string = "uint16"
	Uint32 string = "uint32"
	Uint64 string = "uint64"
	Bool   string = "bool"
	Float  string = "float"
	Double string = "double"
	String string = "string"
	Data   string = "data"
	Server string = "server"
	Enum   string = "enum"
)

var _types = []string{
	Int8, Int16, Int32, Int64, Uint8, Uint16, Uint32, Uint64, Bool, Float, Double, String,
}

var _keys = []string{
	Int8, Int16, Int32, Int64, Uint8, Uint16, Uint32, Uint64, Bool, Float, Double, String, Data, Server, Enum,
}

func Build(out string, in string, typ string) error {
	in = filepath.Clean(in)
	path := filepath.Dir(in)
	name := in[len(path)+1:]
	reg, err := regexp.Compile(strings.ReplaceAll(name, "*", "(.*)"))
	if err != nil {
		return err
	}

	fset := token.NewFileSet() // positions are relative to fset
	pkg := ast.NewPackage()    // positions are relative to fset
	err = parser.ParseDir(fset, pkg, path, reg)
	if err != nil {
		return err
	}

	return nil
}
