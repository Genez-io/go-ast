package main

import (
	"encoding/json"
	"fmt"
	"go/doc"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"

	genezio_parser "gnz-go-ast/parser"
)

type Error struct {
	Error string `json:"error"`
}

func SendError(err error) {
	json, err := json.Marshal(Error{
		Error: err.Error(),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func main() {
	filePath := os.Args[1]
	cwd, err := os.Getwd()
	if err != nil {
		SendError(err)
	}
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedCompiledGoFiles | packages.NeedName | packages.NeedFiles,
		Dir:  cwd,
	}
	classDirectory, err := filepath.Abs(filepath.Dir(filePath))
	if err != nil {
		SendError(err)
	}
	pkgs, err := packages.Load(cfg, classDirectory)
	if err != nil {
		SendError(err)
		return
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			SendError(err)
			return
		}
		p, err := doc.NewFromFiles(pkg.Fset, pkg.Syntax, "")
		if err != nil {
			SendError(err)
			return
		}
		astParser := genezio_parser.New(pkg.TypesInfo, pkg.Types, p)
		err = astParser.Parse(pkg.Syntax[0])
		if err != nil {
			SendError(err)
			return
		}
		json, err := json.MarshalIndent(astParser.Program, "", "  ")
		if err != nil {
			SendError(err)
			return
		}
		fmt.Println(string(json))
	}
}
