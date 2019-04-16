package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
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

func consumeSums(dir string, wq <-chan string, silent bool, ret chan<- *releaseMap, wg *sync.WaitGroup) {
	defer wg.Done()
	for rel := range wq {
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
				if !silent {
					fmt.Printf("Mapped %s\n", path)
				}
			}
		}
		ret <- &relmap
	}
}

func buildSumsMap(dir string, silent bool, jobs int, releases []string) map[string]releaseMap {
	var wg sync.WaitGroup

	relChan := make(chan string, jobs+1)
	mapChan := make(chan *releaseMap, jobs+1)

	relMaps := make(map[string]releaseMap)

	go func(releases []string, ret chan<- string) {
		for _, rel := range releases {
			ret <- rel
		}

		close(relChan)
	}(releases, relChan)

	for workers := 1; workers < jobs; workers++ {
		wg.Add(1)
		go consumeSums(dir, relChan, silent, mapChan, &wg)
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
