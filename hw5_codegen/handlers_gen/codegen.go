package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
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

	caseDefault = `	default:
		w.WriteHeader(http.StatusNotFound)
		data, _ := json.Marshal(resp{"error": "unknown method"})
		w.Write(data)
		return
	}`

	authTempl = `	if at := r.Header.Get("X-Auth"); at != "100500" {
		w.WriteHeader(http.StatusForbidden)
		data, _ := json.Marshal(resp{"error": "unauthorized"})
		w.Write(data)
		return
	}`

	funcsBody = `
	if err != nil {
		if v, ok := err.(ApiError); ok {
			w.WriteHeader(v.HTTPStatus)
			data, _ := json.Marshal(resp{"error": v.Error()})
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		data, _ := json.Marshal(resp{"error": err.Error()})
		w.Write(data)
		return
	}

	response := map[string]interface{}{
		"error":    "",
		"response": user,
	}
	data, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
`
)

type (
	funcMarkData struct {
		URL    string `json:"url"`
		Auth   bool   `json:"auth"`
		Method string `json:"method"`
	}

	funcGenInfo struct {
		astFunc          *ast.FuncDecl
		mark             funcMarkData
		receiverAlias    string
		funcParamsStruct *ast.StructType
		funcStructName   string
	}

	validatorData struct {
		required  bool
		min       string
		max       string
		paramname string
		enum      []string
		sDefault  string
	}
)

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)

	if err != nil {
		log.Fatal(err)
	}
	genInfo := make(map[string][]*funcGenInfo)

	out, _ := os.Create(os.Args[2])
	//out := os.Stdout
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

		for _, comment := range fnc.Doc.List {
			if strings.HasPrefix(comment.Text, "// apigen:api") {
				var mark funcMarkData
				j := strings.TrimLeft(comment.Text, "// apigen:api ")
				if err := json.Unmarshal([]byte(j), &mark); err != nil {
					panic(err)
				}

				if fnc.Recv == nil {
					continue
				}

				if len(fnc.Type.Params.List) < 1 {
					continue
				}

				funcParamStruct, ok := fnc.Type.Params.List[1].Type.(*ast.Ident)
				if !ok {
					continue
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

				for _, f := range node.Decls {
					g, ok := f.(*ast.GenDecl)
					if !ok {
						continue
					}
				SPECS_LOOP:
					for _, spec := range g.Specs {
						currType, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}

						currStruct, ok := currType.Type.(*ast.StructType)
						if !ok {
							continue
						}
						if currType.Name.Name == funcParamStruct.Name {
							funcGen.funcParamsStruct = currStruct
							funcGen.funcStructName = currType.Name.Name
							break SPECS_LOOP
						}
					}
				}
			}
		}
	}

	if len(genInfo) == 0 {
		return
	}

	var serveHTTPBuf bytes.Buffer
	var funcsBuf bytes.Buffer

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
			serveHTTPBuf.WriteString("\tcase \"" + fnc.mark.URL + "\":")
			if fnc.mark.Method == "POST" {
				serveHTTPBuf.WriteString(checkMethodPost)
				serveHTTPBuf.WriteString("\n\t\t")
				serveHTTPBuf.WriteString("srv.handle" + k + fnc.astFunc.Name.Name + "(w, r)\n")
			} else {
				serveHTTPBuf.WriteString(checkMethods)
				serveHTTPBuf.WriteString("\n\t\t")
				serveHTTPBuf.WriteString("srv.handle" + k + fnc.astFunc.Name.Name + "(w, r)\n")
			}
			funcsBuf.WriteString(fmt.Sprintf("func (%s *%s) %s(w http.ResponseWriter, r *http.Request) {\n", fnc.receiverAlias, k, "handle"+k+fnc.astFunc.Name.Name))
			funcsBuf.WriteString("\tw.Header().Set(\"Content-Type\", \"application/json\")\n")

			if fnc.mark.Auth {
				funcsBuf.WriteString(authTempl)
				funcsBuf.WriteString("\n\n")
			}

			if fnc.funcParamsStruct == nil {
				continue
			}

			paramsMap := make(map[string]string)
		FIELDS_LOOP:
			for _, field := range fnc.funcParamsStruct.Fields.List {

				if field.Tag == nil {
					continue
				}
				var tagValue string
				tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
				tagValue = tag.Get("apivalidator")
				if tagValue == "-" {
					continue FIELDS_LOOP
				}

				validStruct := fieldTagParser(tagValue)
				fieldName := field.Names[0].Name
				paramName := strings.ToLower(fieldName)
				fileType := field.Type.(*ast.Ident).Name

				var isString bool
				switch fileType {
				case "int":
					isString = false
				case "string":
					isString = true
				default:
					continue
				}

				if validStruct.paramname != "" {
					paramName = validStruct.paramname
				}

				if isString {
					funcsBuf.WriteString("\t" + paramName + " := r.FormValue(\"" + paramName + "\")\n")
				} else {
					funcsBuf.WriteString("\t" + paramName + "_int" + " := r.FormValue(\"" + paramName + "\")\n")
					funcsBuf.WriteString(fmt.Sprintf("\t%s, err := strconv.Atoi(%s)\n", paramName, paramName+"_int"))
					funcsBuf.WriteString("\tif err != nil {\n")
					writeBadReq(&funcsBuf, paramName+" must be int")
					funcsBuf.WriteString("\t}\n")
					funcsBuf.WriteString("\n")

				}

				if validStruct.required {
					if isString {
						funcsBuf.WriteString("\tif " + paramName + " == \"\" {\n")
					} else {
						funcsBuf.WriteString("\tif len(" + paramName + ") == 0 {\n")
					}
					writeBadReq(&funcsBuf, paramName+" must me not empty")
					funcsBuf.WriteString("\t}\n\n")
				}

				if validStruct.sDefault != "" {
					if isString {
						funcsBuf.WriteString(fmt.Sprintf("\tif %s == \"\" {\n", paramName))
						funcsBuf.WriteString(fmt.Sprintf("\t\t%s = \"%s\"\n", paramName, validStruct.sDefault))
					} else {
						funcsBuf.WriteString("\tif " + paramName + " == 0 {\n")
						funcsBuf.WriteString(fmt.Sprintf("\t\t%s = %s\n", paramName, validStruct.sDefault))
					}
					funcsBuf.WriteString("\t}\n")
					funcsBuf.WriteString("\n")
				}

				if validStruct.min != "" {
					if isString {
						funcsBuf.WriteString("\tif len(" + paramName + ") < " + validStruct.min + " {\n")
						writeBadReq(&funcsBuf, fmt.Sprintf("%s len must be >= %s", paramName, validStruct.min))
					} else {
						funcsBuf.WriteString("\tif " + paramName + " < " + validStruct.min + " {\n")
						writeBadReq(&funcsBuf, fmt.Sprintf("%s must be >= %s", paramName, validStruct.min))
					}
					funcsBuf.WriteString("\t}\n\n")
				}

				if validStruct.max != "" {
					if isString {
						funcsBuf.WriteString("\tif len(" + paramName + ") > " + validStruct.max + " {\n")
						writeBadReq(&funcsBuf, fmt.Sprintf("%s len must be <= %s", paramName, validStruct.max))
					} else {
						funcsBuf.WriteString("\tif " + paramName + " > " + validStruct.max + " {\n")
						writeBadReq(&funcsBuf, fmt.Sprintf("%s must be <= %s", paramName, validStruct.max))
					}
					funcsBuf.WriteString("\t}\n\n")
				}

				if len(validStruct.enum) > 0 {
					if validStruct.sDefault != "" {
						funcsBuf.WriteString("\tif !(")

						var temp, errMsg string
						errMsg = "["
						for _, v := range validStruct.enum {
							errMsg += v + ", "
							temp += fmt.Sprintf("%s == \"%s\" || ", paramName, v)
						}
						temp = strings.TrimSuffix(temp, " || ")
						errMsg = strings.TrimSuffix(errMsg, ", ")
						errMsg = errMsg + "]"
						funcsBuf.WriteString(temp+")")
						funcsBuf.WriteString(" {\n")
						writeBadReq(&funcsBuf, paramName+" must be one of "+errMsg)
						funcsBuf.WriteString("\t}\n\n")
					}
				}

				paramsMap[fieldName] = paramName
			}

			funcsBuf.WriteString("\tin := " + fnc.funcStructName + "{\n")
			for k, v := range paramsMap {
				funcsBuf.WriteString(fmt.Sprintf("\t\t%s: %s,\n", k, v))
			}
			funcsBuf.WriteString("\t}\n\n")
			//	user, err := srv.Create(context.Background(), in)
			// fnc.receiverAlias, k, "handle"+k+fnc.astFunc.Name.Name
			funcsBuf.WriteString(fmt.Sprintf("\tuser, err := %s.%s(context.Background(), in)\n", fnc.receiverAlias, fnc.astFunc.Name.Name))
			funcsBuf.WriteString(funcsBody)
			funcsBuf.WriteString("}\n")

		}
		serveHTTPBuf.WriteString(caseDefault)
		serveHTTPBuf.WriteString("\n")
		serveHTTPBuf.WriteString("}\n")
		serveHTTPBuf.WriteString("\n")
	}
	fmt.Fprintln(out, serveHTTPBuf.String())
	fmt.Fprintln(out, funcsBuf.String())

}

func writeBadReq(buf *bytes.Buffer, errMsg string) {
	buf.WriteString("\t\tw.WriteHeader(http.StatusBadRequest)\n")
	buf.WriteString("\t\tdata, _ := json.Marshal(resp{\"error\": \"" + errMsg + "\"})\n")
	buf.WriteString("\t\tw.Write(data)\n")
	buf.WriteString("\t\treturn\n")
}

func fieldTagParser(s string) validatorData {
	var out validatorData
	data := strings.Split(s, ",")

	for _, d := range data {
		p := strings.Split(d, "=")
		switch p[0] {
		case "required":
			out.required = true
		case "min":
			out.min = p[1]
		case "max":
			out.max = p[1]
		case "paramname":
			out.paramname = p[1]
		case "default":
			out.sDefault = p[1]
		case "enum":
			for _, e := range strings.Split(p[1], "|") {
				out.enum = append(out.enum, e)
			}
		}
	}
	return out
}
