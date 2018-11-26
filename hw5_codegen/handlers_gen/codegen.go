package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"go/ast"
)

// код писать тут

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	//out, _ := os.Create(os.Args[2])
	out := os.Stdout
	fmt.Fprintln(out, `package `+node.Name.Name)

	for _, f := range node.Decls {
		switch f.(type) {
		case *ast.GenDecl:
			g, _ := f.(*ast.GenDecl)
			for _, spec := range g.Specs {
				currType, ok := spec.(*ast.TypeSpec)
				if !ok {
					fmt.Printf("SKIP %T is not ast.TypeSpec\n", spec)
					continue
				}
				currStruct, ok := currType.Type.(*ast.StructType)
				if !ok {
					fmt.Printf("SKIP %T is not ast.StructType\n", currStruct)
					continue
				}
			}
		case *ast.FuncDecl:
			fun, _ := f.(*ast.FuncDecl)
			fmt.Println(fun.Name)
		}
	}
}
