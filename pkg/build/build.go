package build

import "hbuf/pkg/dart"
import "hbuf/pkg/golang"

var buildInits = map[string]func(){
	"dart": dart.Init,
	"go":   golang.Init,
}

func CheckType(typ string) bool {
	_, ok := buildInits[typ]
	return ok
}

func Build(out string, in string, typ string) {

}
