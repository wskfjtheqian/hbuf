package main

import (
	"flag"
	"fmt"
	"hbuf/pkg/build"
	"hbuf/pkg/dart"
	"hbuf/pkg/golang"
	"hbuf/pkg/java"
	ts "hbuf/pkg/typescript"
	"log"
)

var version = "0.0.1"

func main() {
	build.AddBuildType("dart", dart.Build)
	build.AddBuildType("go", golang.Build)
	build.AddBuildType("java", java.Build)
	build.AddBuildType("ts", ts.Build)

	var out = flag.String("o", "", "out dir")
	var in = flag.String("i", "", "input dir")
	var typ = flag.String("t", "", "out type")
	var pack = flag.String("p", "", "package path")
	var showVersion = flag.Bool("v", false, "show version")

	flag.Parse()

	if *showVersion {
		log.Println("Version:", version)
		return
	}

	if nil == out || 0 == len(*out) {
		log.Fatalln("Output directory not found")
	}
	if nil == in || 0 == len(*in) {
		log.Fatalln("Input file not found")
	}
	if nil == pack || 0 == len(*pack) {

	}
	if nil == typ || 0 == len(*typ) {
		log.Fatalln("Build type not found")
	}
	if !build.CheckType(*typ) {
		fmt.Println(fmt.Errorf("Type error : %s", *typ))
		return
	}

	err := build.Build(*out, *in, *typ, *pack)
	if err != nil {
		fmt.Println(fmt.Errorf("Build error: %s", err))
		return
	}
}
