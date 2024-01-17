package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"

	genezio_parser "gnz-go-ast/parser"
)

func main() {
	filePath := os.Args[1]
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedCompiledGoFiles | packages.NeedName | packages.NeedFiles,
		Dir:  cwd,
	}
	classDirectory, err := filepath.Abs(filepath.Dir(filePath))
	if err != nil {
		panic(err)
	}
	pkgs, err := packages.Load(cfg, classDirectory)
	if err != nil {
		panic(err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			panic(err)
		}
		fmt.Println(pkg.Name, pkg.GoFiles)
		astParser := genezio_parser.New(pkg.TypesInfo, pkg.Types)
		err = astParser.Parse(pkg.Syntax[0])
		if err != nil {
			panic(err)
		}
		json, err := json.MarshalIndent(astParser.Program, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(json))
	}
}
