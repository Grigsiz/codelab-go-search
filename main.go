package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	recursiveFlag = flag.Bool("r", false, "recursive search: for directories")
	numLine       = flag.Bool("n", false, "print line number if found")
)

type ScanResult struct {
	file       string
	lineNumber int
	line       string
}

//type rE struct{
//	result []ScanResult
//	err    error
//}

func scanFile(fpath, pattern string) ([]ScanResult, error) {
	f, err := os.Open(fpath)
	ln := 0
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	result := make([]ScanResult, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, pattern) {
			result = append(result, ScanResult{fpath, ln, line})
		}
		ln++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func exit(format string, val ...interface{}) {
	if len(val) == 0 {
		fmt.Println(format)
	} else {
		fmt.Printf(format, val)
		fmt.Println()
	}
	os.Exit(1)
}

func processFile(fpath string, pattern string) chan []ScanResult {
	defer wg.Done()
	var lg sync.WaitGroup
	res := make(chan []ScanResult)
	lg.Add(1)
	go func() {
		defer lg.Done()
		result, err := scanFile(fpath, pattern)
		if err != nil {
			exit("Error scanning %s: %s", fpath, err.Error())
		}
		res <- result
	}()
	go func() {
		lg.Wait()
		close(res)
	}()
	return res
	//res, err := scanFile(fpath, pattern)

}
func Printout(res chan []ScanResult) {
	defer wg.Done()
	numLine := *numLine
	for lines := range res {
		for _, line := range lines {
			if !numLine {
				fmt.Println(line.file, ":", line.line)
			} else {
				fmt.Println(line.file, ":", line.lineNumber, ":", line.line)
			}
		}
	}
}

type FileError struct {
	filename string
	err      error
}

func (fe FileError) Error() string {
	return fe.filename + ":" + fe.err.Error()
}

var wg sync.WaitGroup

func processDirectory(dir string, pattern string) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			fmt.Println("Error:", err, "Path: ", path)
			return err
		}
		if !info.IsDir() {
			wg.Add(2)
			go Printout(processFile(path, pattern))
		}
		return nil
	})
	if err != nil {
		fError, ok := err.(FileError)
		if ok {
			fmt.Println(fError.filename, ":error:", fError.err)
		} else {
			panic(err)
		}
	}
}

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		exit("usage: go-search <path> <pattern> to search")
	}

	path := flag.Arg(0)
	pattern := flag.Arg(1)

	info, err := os.Stat(path)
	if err != nil {
		panic(err)
	}

	recursive := *recursiveFlag
	if info.IsDir() && !recursive {
		exit("%s: is a directory", info.Name())
	}

	if info.IsDir() && recursive {
		processDirectory(path, pattern)
	} else {
		processFile(path, pattern)
	}
	wg.Wait()
}
