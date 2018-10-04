package main

import (
	"os"
	"io"
	"io/ioutil"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

/*
Отступы - символ графики + символ табуляции ( \t )
Для расчета символа графики в отступах подумайте про последний элемент и префикс предыдущих уровней.
Там довольно простое условие. Хорошо помогает проговорить вслух то что вы видите на экране.
 */

func main() {
	out := os.Stdout
	//if !(len(os.Args) == 2 || len(os.Args) == 3) {
	//	panic("usage go run main.go . [-f]")
	//}
	//path := os.Args[1]
	path:= `E:\gopath\src\github.com\coursera\hw1_tree\testdata`
	//printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	printFiles := false
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	return walker(out, path, printFiles, 0)
}

func walker(out io.Writer, path string, printFiles bool, prefixCount int) error {
	list, err := ioutil.ReadDir(path)

	if err != nil {
		return fmt.Errorf("error during open directory %s: %s", path, err.Error())
	}

	for _, v := range list {
		if v.IsDir() {
			prefix := strings.Repeat("│\t", prefixCount)
			fmt.Fprintf(out, prefix + "├───%s\n", v.Name())
			walker(out, filepath.Join(path, v.Name()), printFiles, prefixCount+1)
		} else {
			prefix := strings.Repeat("│\t", prefixCount)
			fmt.Fprintf(out, prefix + "├───%s (%s)\n", v.Name(), getFileSize(v))
		}
	}
	return nil
}

func getFileSize(fileInfo os.FileInfo) string {
	size := fileInfo.Size()
	if size == 0 {
		return "empty"
	}
	return strconv.Itoa(int(size)) + "b"
}

