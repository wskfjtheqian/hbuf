package build

import (
	"hbuf/pkg/dart"
	"hbuf/pkg/parser"
	"hbuf/pkg/scanner"
	"hbuf/pkg/token"
	"io/fs"
	"io/ioutil"
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

func Build(out string, in string, typ string) error {
	in = filepath.Clean(in)
	path := filepath.Dir(in)
	name := in[len(path)+1:]
	reg, err := regexp.Compile(strings.ReplaceAll(name, "*", "(.*)"))
	if err != nil {
		return err
	}

	fset := token.NewFileSet() // positions are relative to fset
	err = buildDir(fset, path, reg)
	if err != nil {
		return err
	}

	return nil
}

func buildDir(fset *token.FileSet, path string, reg *regexp.Regexp) error {
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, item := range dir {
		if item.IsDir() {
			err := buildDir(nil, filepath.Join(path, item.Name()), reg)
			if err != nil {
				return err
			}
		}
		if !reg.MatchString(item.Name()) {
			continue
		}
		err := parseFile(fset, path, item.Name())
		if err != nil {
			return err
		}
	}
	return nil
}

func parseFile(fset *token.FileSet, path string, name string) error {
	file := filepath.Join(path, name)
	if nil != fset.GetFileByName(file) {
		return nil
	}
	f, err := parser.ParseFile(fset, file, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return err
	}
	for _, spec := range f.Imports {
		imp := spec.Path.Value
		imp = imp[1 : len(imp)-1]
		if !("/" == imp || 0 != len(filepath.VolumeName(imp))) {
			imp = filepath.Join(path, imp)
		}
		path := filepath.Dir(imp)
		name := imp[len(path)+1:]
		err := parseFile(fset, path, name)
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
