package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const (
	tabVerticalLine = "│\t"
	tab             = "\t"
	dirPrefix       = "├───"
	lastDirPrefix   = "└───"
)

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
	return walker(out, path, printFiles, "")
}

func walker(out io.Writer, path string, printFiles bool, prefix string) error {
	var lastElementIndex int
	list, err := ioutil.ReadDir(path)
	listLen := len(list)
	lastElementIndex = listLen - 1

	if !printFiles {
		for i := listLen - 1; i >= 0; i-- {
			if list[i].IsDir() {
				lastElementIndex = i
				break
			}
		}
	}

	if err != nil {
		return fmt.Errorf("error during open directory %s: %s", path, err.Error())
	}

	for i, v := range list {
		var newPrefix, outPrefix string
		outPrefix = prefix + dirPrefix
		newPrefix = prefix + tabVerticalLine

		if i == lastElementIndex {
			outPrefix = prefix + lastDirPrefix
			newPrefix = prefix + tab
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
