package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
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
	path := `D:\gopath\src\coursera\hw1_tree\testdata`
	//printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	printFiles := true
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

	mainPrefix := strings.Repeat("│\t", prefixCount)
	mainLastPrefix := "│" + strings.Repeat("\t", prefixCount)
	dirPreifx := mainPrefix + "├───%s\n"
	dirLastPreifx := mainLastPrefix + "└───%s\n"
	fileRefix := mainPrefix + "├───%s (%s)\n"
	fileLastRefix := mainLastPrefix + "└───%s (%s)\n"
	lastElementIndex := len(list) - 1

	for i, v := range list {
		if v.IsDir() {
			if i == lastElementIndex {
				fmt.Fprintf(out, dirLastPreifx, v.Name())
			} else {
				fmt.Fprintf(out, dirPreifx, v.Name())
			}
			walker(out, filepath.Join(path, v.Name()), printFiles, prefixCount+1)
		} else if printFiles {
			if i == lastElementIndex {
				fmt.Fprintf(out, fileLastRefix, v.Name(), getFileSize(v))
			} else {
				fmt.Fprintf(out, fileRefix, v.Name(), getFileSize(v))
			}
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
