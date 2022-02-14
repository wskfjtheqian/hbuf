package build

import (
	"hbuf/pkg/dart"
	"hbuf/pkg/parser"
	"hbuf/pkg/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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
	stat, err := os.Stat(in)
	if err != nil {
		return err
	}
	reg, err := regexp.Compile(stat.Name())
	if err != nil {
		return err
	}

	fset := token.NewFileSet() // positions are relative to fset
	err = buildDir(fset, filepath.Dir(in), reg)
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
		_, err := parser.ParseFile(fset, filepath.Join(path, item.Name()), nil, parser.AllErrors|parser.ParseComments)
		if err != nil {
			return err
		}
	}
	return nil
}
