package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"encoding/json"
)

var (
	checkMethods = `
	if !(r.Method == http.MethodGet || r.Method == http.MethodPost) {
		w.WriteHeader(http.StatusNotAcceptable)
		data, _ := json.Marshal(resp{"error":"bad method"})
		w.Write(data)
		return
	}
`
	checkMethodPost = `
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotAcceptable)
		data, _ := json.Marshal(resp{"error": "bad method"})
		w.Write(data)
		return
	}
`
)

type funcMarkData struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

func main() {
	fset := token.NewFileSet()
	//node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	node, err := parser.ParseFile(fset, "api.go", nil, parser.ParseComments)

	if err != nil {
		log.Fatal(err)
	}

	//out, _ := os.Create(os.Args[2])
	out := os.Stdout
	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `import "context"`)
	fmt.Fprintln(out, `import "encoding/json"`)
	fmt.Fprintln(out, `import "net/http"`)
	fmt.Fprintln(out, `import "strconv"`)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `type resp map[string]interface{}`)

	for _, f := range node.Decls {
		fnc, ok := f.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fnc.Doc == nil {
			//fmt.Printf("SKIP func %#v doesnt have comments\n", fnc.Name.Name)
			continue
		}
		needCodegen := false
		var mark funcMarkData
		for _, comment := range fnc.Doc.List {
			if strings.HasPrefix(comment.Text, "// apigen:api") {
				needCodegen = true
				j := strings.TrimLeft(comment.Text, "// apigen:api ")
				if err := json.Unmarshal([]byte(j), &mark); err != nil {
					panic(err)
				}
			}
		}
		if !needCodegen {
			//fmt.Printf("SKIP func %#v doesnt have apigen:api mark\n", fnc.Name.Name)
			continue
		}
		fmt.Printf("type: %T data: %+v\n", fnc.Type.Params.List[1], fnc.Type.Params.List[1])
	}
}
