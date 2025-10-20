package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
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
	printFilesInDir(out, path, "", printFiles)
	return nil
}

func getFileSize(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func getFileSizeStr(path string) string {

	fi, _ := os.Stat(path)

	if fi.IsDir() {
		return ""
	}

	bytes, err := getFileSize(path)
	if err != nil {
		panic(err)
	}
	if bytes == 0 {
		return " (empty)"
	}
	return fmt.Sprintf(" (%db)", bytes)
}

func printFilesInDir(out io.Writer, dir, prefix string, printFiles bool) {
	initFiles, _ := os.ReadDir(dir)

	files := make([]os.DirEntry, 0, len(initFiles))

	for _, file := range initFiles {
		if file.IsDir() {
			files = append(files, file)
		} else if printFiles {
			files = append(files, file)
		}
	}

	for i, file := range files {

		name := file.Name()
		parent := filepath.Join(dir, name)
		branch := "├"
		newPrefix := prefix + "│\t"

		if i == len(files)-1 {
			branch = "└"
			newPrefix = prefix + "\t"
		}

		if !file.IsDir() && printFiles {
			name = name + getFileSizeStr(parent)
		}

		out.Write([]byte(fmt.Sprintf("%s%s───%s\n", prefix, branch, name)))

		printFilesInDir(out, path.Join(dir, file.Name()), newPrefix, printFiles)

	}
}
