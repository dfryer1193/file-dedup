package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

func getReleases(dir, rel string) ([]string, error) {
	var releases []string
	r := regexp.MustCompile(`.*\.sums$`)

	if rel != "0" {
		r = regexp.MustCompile(rel + `.*\.sums$`)
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
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

func needsDedup(releases []string, maps map[string]map[string]string) (map[string][]string, error) {
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

func dedup(dir string, dups map[string][]string) error {
	for file, rels := range dups {
		sort.Sort(byRelease(rels))
		latestRel := rels[len(rels)-1]

		file := strings.TrimPrefix(file, ".")

		for _, rel := range rels {

			if rel == latestRel {
				continue
			}

			latestFilePath := dir + latestRel + file
			oldFilePath := dir + rel + file

			latestFInfo, err := os.Stat(latestFilePath)
			if err != nil {
				fmt.Printf("Source file %s does not exist!\n", latestFilePath)
				continue
			}

			fInfo, err := os.Stat(oldFilePath)
			if err != nil {
				fmt.Printf("File %s does not exist!\n", oldFilePath)
				continue
			}

			if os.SameFile(latestFInfo, fInfo) {
				fmt.Printf("Skipping identical file %s\n", fInfo.Name())
				continue
			}

			err = os.Rename(oldFilePath, oldFilePath+".bak")
			if err != nil {
				fmt.Printf("Could not backup old file %s\n", oldFilePath)
				continue
			}

			err = os.Link(latestFilePath, oldFilePath)
			if err != nil {
				fmt.Printf("Failed to link file %s to %s\n", latestFilePath, oldFilePath)
				err2 := os.Rename(oldFilePath+".bak", oldFilePath)
				if err2 != nil {
					fmt.Printf("Could not rename file %s.bak to %s\n", oldFilePath, oldFilePath)
				}
				continue
			}

			err = os.Remove(oldFilePath + ".bak")
			if err != nil {
				fmt.Printf("Failed to remove backup file %s.bak\n", oldFilePath)
			}
		}
	}
	return nil
}

func main() {
	var release, directory string

	flag.StringVar(&release, "rel", "0", "The release to dedup. 0 for all releases.")
	flag.StringVar(&directory, "dir", "./", "The directory containing releases to dedup.")
	flag.Parse()

	if !strings.HasSuffix(directory, "/") {
		directory = directory + "/"
	}

	releases, err := getReleases(directory, release)
	if err != nil {
		log.Fatal(err)
	}

	if len(releases) < 2 {
		fmt.Printf("Not enough releases to dedup! Exiting...\n")
		os.Exit(0)
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

	dups, err := needsDedup(releases, maps)
	if err != nil {
		log.Fatal(err)
	}

	err = dedup(directory, dups)
	if err != nil {
		log.Fatal(err)
	}
}
