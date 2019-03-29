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
		r := regexp.MustCompile(`.*\.sums$`)
		matched := r.MatchString(f.Name())
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
	var err error

	maps := make(map[string]map[string]string)

	consume := func(fname string) (map[string]string, error) {
		fileHashes := make(map[string]string)

		file, err := os.Open(fname)
		if err != nil {
			return nil, err
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
				fileHashes[path] = hash
			}
		}

		return fileHashes, nil
	}

	for _, rel := range releases {
		maps[rel], err = consume(dir + rel + ".sums")
		if err != nil {
			return nil, err
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
		for path, hash := range latestMap {
			if relMap[path] == hash {
				dups[path] = append(dups[path], rel)
			}
		}
	}

	return dups, nil
}

func dedup(dups map[string][]string) error {
	for file, rels := range dups {
		if len(rels) == 1 {
			continue
		}

		sort.Sort(byRelease(rels))
		latestRel := rels[len(rels)-1]
		for i, rel := range rels {
			if i == len(rels)-1 {
				fmt.Printf("Skipping newest release %s...\n", rel)
				continue
			}

			fmt.Printf("mv %s/%s{,~}\n"+
				"ln %s/%s %s/%s\n"+
				"$? && rm %s/%s~\n\n",
				rel, file,
				latestRel, file, rel, file,
				rel, file)
		}
	}
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

	//for path, rels := range dups {
	//	fmt.Printf("%s: %v\n", path, rels)
	//}

	dedup(dups)
}
