package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

func getReleases(dir string) ([]string, error) {
	var releases []string

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		matched, err := regexp.MatchString(`.*\.sums$`, f.Name())
		if err != nil {
			return nil, err
		}

		if matched {
			release := strings.TrimSuffix(f.Name(), ".sums")
			releases = append(releases, release)
		}
	}

	return releases, nil
}

func buildFileMap(dir string, releases []string) (map[string]map[string]string, error) {
	maps := make(map[string]map[string]string)

	for _, rel := range releases {
		file, err := os.Open(dir + rel + ".sums")
		if err != nil {
			return nil, err
		}
		defer file.Close()

		maps[rel] = make(map[string]string)

		scanner := bufio.NewScanner(bufio.NewReader(file))
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			splitLine := strings.SplitN(scanner.Text(), " ", 2)
			if len(splitLine) < 2 {
				fmt.Printf("Cannot parse line: %s\n", splitLine)
			} else {
				path := strings.TrimSpace(splitLine[1])
				hash := splitLine[0]
				maps[rel][path] = hash
			}
		}
	}
	return maps, nil
}

func needsDedup(dir string, releases []string, maps map[string]map[string]string) (map[string][]string, error) {
	dups := make(map[string][]string)

	latestRelease := releases[len(releases)-1]
	latestMap := maps[latestRelease]

	if latestMap == nil {
		return nil, fmt.Errorf("map %s does not exist", latestRelease)
	}

	for rel, relMap := range maps {
		if rel == latestRelease {
			continue
		}

		for path, hash := range latestMap {
			if relMap[path] == hash {
				dups[path] = append(dups[path], rel)
			}
		}
	}

	return dups, nil
}

func dedup(dups map[string][]string) error {
	// TODO: Implement deduplication
	return nil
}

func main() {
	directory := "./"
	if 1 < len(os.Args) {
		directory = os.Args[1]
	}

	releases, err := getReleases(directory)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Split by release
	sort.Sort(byRelease(releases))

	maps, err := buildFileMap(directory, releases)
	if err != nil {
		log.Fatal(err)
	}

	if len(maps) == 0 {
		fmt.Println("No Map Built!")
	}

	dups, err := needsDedup(directory, releases, maps)
	if err != nil {
		log.Fatal(err)
	}

	for path, rels := range dups {
		fmt.Printf("%s: %v\n", path, rels)
	}

	dedup(dups)
}
