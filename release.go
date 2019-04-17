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

func getReleases(dir, rel string, useSums, silent bool) ([]string, error) {
	var releases []string

	regexStr := releaseRegex
	if useSums {
		regexStr = sumsRegex
	}

	if rel != "0" {
		regexStr = rel + regexStr
	} else {
		regexStr = `[0-9]+` + regexStr
	}

	r := regexp.MustCompile(regexStr)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		matched := r.MatchString(f.Name())

		if matched {
			release := f.Name()
			if useSums {
				release = strings.TrimSuffix(release, ".sums")
			}

			if !silent {
				fmt.Printf("Found release %s\n", release)
			}

			releases = append(releases, release)
		}
	}

	return releases, nil
}

func consumeSums(sumsFile string, silent bool) map[string]string {
	relmap := make(map[string]string)

	file, err := os.Open(sumsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(bufio.NewReader(file))
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		splitLine := strings.SplitN(scanner.Text(), " ", 2)
		if len(splitLine) < 2 {
			fmt.Printf("Cannot parse line: %s\n", splitLine)
		} else {
			path := strings.TrimSpace(splitLine[1])
			hash := splitLine[0]
			relmap[path] = hash
			if !silent {
				fmt.Printf("Mapped %s\n", path)
			}
		}
	}
	return relmap
}

func buildSumsMap(dir string, silent bool, jobs int, releases []string) map[string]map[string]string {
	relMaps := make(map[string]map[string]string)
	for _, rel := range releases {
		fname := dir + rel + ".sums"
		relMaps[rel] = make(map[string]string)
		relMaps[rel] = consumeSums(fname, silent)
	}
	return relMaps
}
