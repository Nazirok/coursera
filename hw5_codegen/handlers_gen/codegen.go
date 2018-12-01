package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"bytes"
)

var (
	checkMethods = `
		if !(r.Method == http.MethodGet || r.Method == http.MethodPost) {
			w.WriteHeader(http.StatusNotAcceptable)
			data, _ := json.Marshal(resp{"error":"bad method"})
			w.Write(data)
			return
		}`

	checkMethodPost = `
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotAcceptable)
			data, _ := json.Marshal(resp{"error": "bad method"})
			w.Write(data)
			return
		}`

	caseDefault =
`	default:
		w.WriteHeader(http.StatusNotFound)
		data, _ := json.Marshal(resp{"error": "unknown method"})
		w.Write(data)
		return
	}`
)

type (
	funcMarkData struct {
		URL    string `json:"url"`
		Auth   bool   `json:"auth"`
		Method string `json:"method"`
	}

	funcGenInfo struct {
		astFunc       *ast.FuncDecl
		mark          funcMarkData
		receiverAlias string
	}
)

func main() {
	fset := token.NewFileSet()
	//node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	node, err := parser.ParseFile(fset, "api.go", nil, parser.ParseComments)

	if err != nil {
		log.Fatal(err)
	}
	genInfo := make(map[string][]*funcGenInfo)

	//out, _ := os.Create(os.Args[2])
	out := os.Stdout
	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out)

	for _, f := range node.Decls {
		fnc, ok := f.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fnc.Doc == nil {
			continue
		}

		//var mark funcMarkData
		for _, comment := range fnc.Doc.List {
			if strings.HasPrefix(comment.Text, "// apigen:api") {
				if fnc.Recv == nil {
					continue
				}

				var mark funcMarkData
				j := strings.TrimLeft(comment.Text, "// apigen:api ")
				if err := json.Unmarshal([]byte(j), &mark); err != nil {
					panic(err)
				}

				recv := fnc.Recv.List[0]
				funcGen := &funcGenInfo{
					mark:          mark,
					astFunc:       fnc,
					receiverAlias: recv.Names[0].Name,
				}

				v, ok := recv.Type.(*ast.StarExpr)
				if ok {
					x, ok := v.X.(*ast.Ident)
					if ok {
						genInfo[x.Name] = append(genInfo[x.Name], funcGen)
					}
				}
			}
		}
	}

	if len(genInfo) == 0 {
		return
	}

	var serveHTTPBuf bytes.Buffer
	//var funcsBuf bytes.Buffer

	fmt.Fprintln(out, `import "context"`)
	fmt.Fprintln(out, `import "encoding/json"`)
	fmt.Fprintln(out, `import "net/http"`)
	fmt.Fprintln(out, `import "strconv"`)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `type resp map[string]interface{}`)
	fmt.Fprintln(out)

	for k, v := range genInfo {
		serveHTTPBuf.WriteString(fmt.Sprintf("func (%s *%s) ServeHTTP(w http.ResponseWriter, r *http.Request) {\n", v[0].receiverAlias, k))
		serveHTTPBuf.WriteString("\t")
		serveHTTPBuf.WriteString("switch r.URL.Path {\n")
		for _, fnc := range v {
			serveHTTPBuf.WriteString("\tcase \""+fnc.mark.URL+"\":")
			if fnc.mark.Method == "POST" {
				serveHTTPBuf.WriteString(checkMethodPost)
				serveHTTPBuf.WriteString("\n\t\t")
				serveHTTPBuf.WriteString("srv.handle"+k+fnc.astFunc.Name.Name+"(w, r)\n")
			} else {
				serveHTTPBuf.WriteString(checkMethods)
				serveHTTPBuf.WriteString("\n\t\t")
				serveHTTPBuf.WriteString("srv.handle"+k+fnc.astFunc.Name.Name+"(w, r)\n")
			}
		}
		serveHTTPBuf.WriteString(caseDefault)
		serveHTTPBuf.WriteString("\n")

	}
	fmt.Fprintln(out, serveHTTPBuf.String())
}
