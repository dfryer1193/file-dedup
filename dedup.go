package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type releaseMap struct {
	release string
	fileMap map[string]string
}

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

func consumeRelease(dir, rel string, ret chan<- *releaseMap, wg *sync.WaitGroup) {
	defer wg.Done()
	relmap := releaseMap{
		release: rel,
		fileMap: make(map[string]string),
	}
	fname := dir + rel + ".sums"

	file, err := os.Open(fname)
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
			relmap.fileMap[path] = hash
		}
	}
	ret <- &relmap
}

func buildReleaseMap(dir string, releases []string) map[string]releaseMap {
	const maxChanDepth = 5

	var mapChan chan *releaseMap
	var wg sync.WaitGroup

	if len(releases) < maxChanDepth {
		mapChan = make(chan *releaseMap, len(releases))
	} else {
		mapChan = make(chan *releaseMap, maxChanDepth)
	}

	relMaps := make(map[string]releaseMap)

	for _, rel := range releases {
		go consumeRelease(dir, rel, mapChan, &wg)
		wg.Add(1)
	}

	for i := len(releases); i > 0; i-- {
		select {
		case relMap := <-mapChan:
			relMaps[relMap.release] = *relMap
		case <-time.After(30 * time.Second):
			log.Fatal(fmt.Errorf("Waiting for map longer than %d seconds", 30))
		}
	}

	wg.Wait()
	close(mapChan)

	return relMaps
}

func needsDedup(releases []string, releaseMaps map[string]releaseMap) (map[string][]string, error) {
	dups := make(map[string][]string)

	latestRelease := releases[len(releases)-1]
	latestMap := releaseMaps[latestRelease]

	if latestMap.fileMap == nil {
		return nil, fmt.Errorf("map %s does not exist", latestRelease)
	}

	for rel, relMap := range releaseMaps {
		for path, hash := range latestMap.fileMap {
			if relMap.fileMap[path] == hash {
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
	var jobs int

	flag.StringVar(&release, "rel", "0", "The release to dedup. 0 for all releases.")
	flag.StringVar(&directory, "dir", "./", "The directory containing releases to dedup.")
	flag.IntVar(&jobs, "j", runtime.NumCPU(), "Maximum number of jobs.")
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

	releaseMaps := buildReleaseMap(directory, releases)

	if len(releaseMaps) == 0 {
		fmt.Println("No Map Built!")
	}

	dups, err := needsDedup(releases, releaseMaps)

	err = dedup(directory, dups)
	if err != nil {
		log.Fatal(err)
	}
}
