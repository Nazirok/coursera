package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"io/ioutil"
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

func dirTree(out io.Writer, basePath string, pintFiles bool) error {

}

func walk(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error during open directory %s: %s", path, err.Error())
	}
	list, _ := f.Readdir(0)
	f.Close()
	sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
	for _, v := range list {
		if v.IsDir() {
			fmt.Printf("├───%s\n", v.Name())
			walk(filepath.Join(path, v.Name()))
		} else {
			fmt.Printf("├───\t%s\n", v.Name())
		}
	}
	return nil
}

func walk2(path string) error {
	list, _ := ioutil.ReadDir(path)
	for _, v := range list {
		if v.IsDir() {
			fmt.Printf("├───%s\n", v.Name())
			walk(filepath.Join(path, v.Name()))
		} else {
			fmt.Printf("├───\t%s\n", v.Name())
		}
	}
	return nil
}
