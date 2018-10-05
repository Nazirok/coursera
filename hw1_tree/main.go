package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

/*
Отступы - символ графики + символ табуляции ( \t )
Для расчета символа графики в отступах подумайте про последний элемент и префикс предыдущих уровней.
Там довольно простое условие. Хорошо помогает проговорить вслух то что вы видите на экране.
*/

const (
	tabVerticalLine = "│\t"
	tab             = "\t"
	dirPrefix       = "├───"
	lastDirPrefix   = "└───"
)

func main() {
	out := os.Stdout
	//if !(len(os.Args) == 2 || len(os.Args) == 3) {
	//	panic("usage go run main.go . [-f]")
	//}
	//path := os.Args[1]
	path := `E:\gopath\src\github.com\coursera\hw1_tree\testdata`
	//printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	printFiles := false
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	return walker(out, path, printFiles, "")
}

func walker(out io.Writer, path string, printFiles bool, prefix string) error {
	list, err := ioutil.ReadDir(path)

	if err != nil {
		return fmt.Errorf("error during open directory %s: %s", path, err.Error())
	}

	//mainPrefix := strings.Repeat("│\t", prefixCount)
	//mainLastPrefix := "│" + strings.Repeat("\t", prefixCount)
	//dirPreifx := prefix + "├───%s\n"
	//dirLastPreifx := prefix + "└───%s\n"
	//fileRefix := prefix + "├───%s (%s)\n"
	//fileLastRefix := prefix + "└───%s (%s)\n"
	lastElementIndex := len(list) - 1

	for i, v := range list {
		var newPrefix, outPrefix string
		if i == lastElementIndex {
			outPrefix = prefix + lastDirPrefix
			newPrefix = prefix + tab
		} else {
			outPrefix = prefix + dirPrefix
			newPrefix = prefix + tabVerticalLine
		}
		if v.IsDir() {
			fmt.Fprintf(out, outPrefix+"%s\n", v.Name())
			walker(out, filepath.Join(path, v.Name()), printFiles, newPrefix)

		} else if printFiles {
			fmt.Fprintf(out, outPrefix+"%s (%s)\n", v.Name(), getFileSize(v))
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
