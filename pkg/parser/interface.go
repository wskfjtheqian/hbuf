package parser

import (
	"bytes"
	"errors"
	"hbuf/pkg/ast"
	"hbuf/pkg/scanner"
	"hbuf/pkg/token"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			return io.ReadAll(s)
		}
		return nil, errors.New("invalid source")
	}
	return os.ReadFile(filename)
}

type Mode uint

const (
	PackageClauseOnly Mode             = 1 << iota // stop parsing after package clause
	ImportsOnly                                    // stop parsing after import declarations
	ParseComments                                  // parse comments and add them to AST
	Trace                                          // print a trace of parsed productions
	DeclarationErrors                              // report declaration errors
	SpuriousErrors                                 // same as AllErrors, for backward-compatibility
	AllErrors         = SpuriousErrors             // report all errors (not just the first 10 on different lines)
)

func ParseFile(fset *token.FileSet, filename string, src interface{}, mode Mode) (f *ast.File, err error) {
	if fset == nil {
		panic("parser.ParseFile: no token.FileSet provided (fset == nil)")
	}

	// get source
	text, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	var p parser
	defer func() {
		if e := recover(); e != nil {
			// resume same panic if it's not a bailout
			if _, ok := e.(bailout); !ok {
				panic(e)
			}
		}

		// set result values
		if f == nil {
			// source is not a valid Go source file - satisfy
			// ParseFile API and return a valid (but) empty
			// *ast.File
			f = &ast.File{
				Scope: ast.NewScope(nil),
			}
		}

		p.errors.Sort()
		err = p.errors.Err()
	}()

	// parse source
	p.init(fset, filename, text, mode)
	f = p.parseFile()

	return
}

func ParseDir(fset *token.FileSet, pkg *ast.Package, path string, reg *regexp.Regexp) error {
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, item := range dir {
		if item.IsDir() {
			err := ParseDir(fset, pkg, filepath.Join(path, item.Name()), reg)
			if err != nil {
				return err
			}
		}
		if !reg.MatchString(item.Name()) {
			continue
		}
		err := parseDirFile(fset, pkg, path, item.Name())
		if err != nil {
			return err
		}
	}
	return nil
}

func parseDirFile(fset *token.FileSet, pkg *ast.Package, path string, name string) error {
	filePath := filepath.Join(path, name)
	var temp = false
	fset.Iterate(func(file *token.File) bool {
		if file.Name() == filePath {
			temp = true
			return false
		}
		return true
	})
	if temp {
		return nil
	}

	f, err := ParseFile(fset, filePath, nil, AllErrors|ParseComments)
	if err != nil {
		return err
	}
	pkg.Files[filePath] = f
	for _, spec := range f.Imports {
		imp := spec.Path.Value
		imp = imp[1 : len(imp)-1]
		if !("/" == imp || 0 != len(filepath.VolumeName(imp))) {
			imp = filepath.Join(path, imp)
			spec.Path.Value = imp
		}
		path := filepath.Dir(imp)
		name := imp[len(path)+1:]
		err := parseDirFile(fset, pkg, path, name)
		if err != nil {
			switch err.(type) {
			case *fs.PathError:
				return scanner.Error{
					Pos: fset.Position(spec.Pos()),
					Msg: "Not import: " + spec.Path.Value,
				}
			}
			return err
		}
	}
	return nil
}
