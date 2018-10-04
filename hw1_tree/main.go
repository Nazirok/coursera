package main

import (
	"os"
	"io"
	"io/ioutil"
	"fmt"
	"path/filepath"
	"strconv"
)

/*
Отступы - символ графики + символ табуляции ( \t )
Для расчета символа графики в отступах подумайте про последний элемент и префикс предыдущих уровней.
Там довольно простое условие. Хорошо помогает проговорить вслух то что вы видите на экране.
 */

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	list, err := ioutil.ReadDir(path)

	if err != nil {
		return fmt.Errorf("error during open directory %s: %s", path, err.Error())
	}

	for _, v := range list {
		if v.IsDir() {
			fmt.Fprintf(out, "├───%s\n", v.Name())
			dirTree(out, filepath.Join(path, v.Name()), printFiles)
		} else {
			fmt.Fprintf(out, "│\t├───%s (%s)\n", v.Name(), getFileSize(v))
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

