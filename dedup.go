package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
)

var releaseRegex = `\.[0-9]+-(GA|RC-[0-9]+|Snap.*-[0-9]+|Beta|Alpha)$`
var sumsRegex = `.*\.sums$`

func buildReleaseMap(dir string, useSums, silent bool, jobs int, releases []string) map[string]map[string]string {
	if useSums {
		return buildSumsMap(dir, silent, jobs, releases)
	}

	return buildHashMap(dir, silent, jobs, releases)
}

func needsDedup(releases []string, releaseMaps map[string]map[string]string) (map[string][]string, error) {
	dups := make(map[string][]string)

	latestRelease := releases[len(releases)-1]
	latestMap := releaseMaps[latestRelease]

	if latestMap == nil {
		return nil, fmt.Errorf("map %s does not exist", latestRelease)
	}

	for rel, relMap := range releaseMaps {
		for path, hash := range latestMap {
			if relMap[path] == hash {
				dups[path] = append(dups[path], rel)
			}
		}
	}
	return dups, nil
}

func dedup(dir string, silent bool, dups map[string][]string) error {
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

			if !silent {
				fmt.Printf("Deduplicating %s\n", oldFilePath)
			}

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
	var jobs int
	var useSums, silent, dryRun bool

	flag.StringVar(&release, "rel", "0", "The release to dedup. 0 for all releases.")
	flag.StringVar(&directory, "dir", "./", "The directory containing releases to dedup.")
	flag.IntVar(&jobs, "j", runtime.NumCPU(), "Maximum number of jobs.")
	flag.BoolVar(&useSums, "sums", false, "Use premade *.sums files")
	flag.BoolVar(&silent, "s", false, "Run silently")
	flag.BoolVar(&dryRun, "d", false, "Do not perform deduplication.")
	flag.Parse()

	if !strings.HasSuffix(directory, "/") {
		directory = directory + "/"
	}

	releases, err := getReleases(directory, release, useSums, silent)
	if err != nil {
		log.Fatal(err)
	}

	if len(releases) < 2 {
		fmt.Printf("Not enough releases to dedup! Exiting...\n")
		os.Exit(0)
	}

	sort.Sort(byRelease(releases))

	releaseMaps := buildReleaseMap(directory, useSums, silent, jobs, releases)

	if len(releaseMaps) == 0 {
		fmt.Println("No Map Built!")
	}

	dups, err := needsDedup(releases, releaseMaps)

	if !dryRun {
		err = dedup(directory, silent, dups)
		if err != nil {
			log.Fatal(err)
		}
	}
}
