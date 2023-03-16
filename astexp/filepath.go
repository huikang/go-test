package astexp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func walkDir(dir string) {
	count := 0
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fmt.Println(path, info.Size())
			if strings.Contains(info.Name(), "_test.go") {
				count++
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	log.Println("total number of file", count)
}

func listTestFiles(dir string) []string {
	results := []string{}
	count := 0
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fmt.Println(path, info.Size())
			if strings.Contains(info.Name(), "_test.go") {
				count++
				results = append(results, path)
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	log.Println("total number of file", count)
	return results
}
