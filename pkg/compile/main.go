package main

import (
	"flag"
	"fmt"
	"hbuf/pkg/build"
	"hbuf/pkg/dart"
	"hbuf/pkg/golang"
)

func main() {
	build.AddBuildType("dart", dart.Build)
	build.AddBuildType("go", golang.Build)

	var out = flag.String("o", "", "out dir")
	var in = flag.String("i", "", "input dir")
	var typ = flag.String("t", "", "out type")
	flag.Parse()

	if nil == out || 0 == len(*out) {

	}
	if nil == in || 0 == len(*in) {

	}
	if nil == typ || 0 == len(*typ) {
		fmt.Println(fmt.Errorf("Not find type"))
		return
	}
	if !build.CheckType(*typ) {
		fmt.Println(fmt.Errorf("Type error : %s", *typ))
		return
	}

	err := build.Build(*out, *in, *typ)
	if err != nil {
		fmt.Println(fmt.Errorf("Build error: %s", err))
		return
	}
}
