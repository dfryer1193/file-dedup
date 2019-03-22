package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

func buildFileMap(dir string) (map[string]map[string]string, error) {
	maps := make(map[string]map[string]string)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		matched, _ := regexp.MatchString(`.*\.sums`, f.Name())
		if matched {
			file, err := os.Open(f.Name())
			if err != nil {
				return nil, err
			}
			defer file.Close()

			release := strings.TrimSuffix(f.Name(), ".sums")

			maps[release] = make(map[string]string)

			scanner := bufio.NewScanner(bufio.NewReader(file))
			scanner.Split(bufio.ScanLines)

			for scanner.Scan() {
				splitLine := strings.SplitN(scanner.Text(), " ", 2)
				if len(splitLine) < 2 {
					fmt.Printf("Cannot parse line: %s\n", splitLine)
				} else {
					path := strings.TrimSpace(splitLine[1])
					hash := splitLine[0]
					maps[release][path] = hash
				}
			}
		}
	}
	return maps, nil
}

func getLatest(maps map[string]map[string]string, majVer int) map[string]string {
	return nil
}

func dedup(map[string]map[string]string) {
}

func main() {
	directory := "./"
	if 1 < len(os.Args) {
		directory = os.Args[1]
	}

	maps, err := buildFileMap(directory)
	if err != nil {
		log.Fatal(err)
	}

	if len(maps) == 0 {
		fmt.Println("No Map Built!")
	}
}
