package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

var mux sync.Mutex

func buildHashMap(dir string, silent bool, jobs int, releases []string) map[string]map[string]string {
	relMaps := make(map[string]map[string]string)

	for _, rel := range releases {
		relmap := make(map[string]string)
		relmap = hashRelease(dir, rel, silent, jobs)
		relMaps[rel] = relmap
	}

	return relMaps
}

func hashRelease(dir, release string, silent bool, jobs int) map[string]string {
	var wg sync.WaitGroup
	wq := make(chan string, jobs+1)
	hashedRelease := make(map[string]string)

	wg.Add(1)
	go getFiles(dir+release+"/", wq, &wg)

	for workers := 1; workers < jobs; workers++ {
		wg.Add(1)
		go hashFile(dir, release, hashedRelease, silent, wq, &wg)
	}

	wg.Wait()

	return hashedRelease
}

func getFiles(dir string, wq chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	searchFiles(dir, wq)
	close(wq)
}

func searchFiles(dir string, wq chan<- string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			searchFiles(dir+file.Name()+"/", wq)
		} else {
			wq <- (dir + file.Name())
		}
	}
}

func hashFile(dir, release string, hashMap map[string]string, silent bool, wq <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	h := sha256.New()

	getSum := func(path string, h hash.Hash) string {
		f, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		io.Copy(h, f)

		return hex.EncodeToString(h.Sum(nil))
	}

	for fpath := range wq {
		if !silent {
			fmt.Printf("Hashing %s\n", fpath)
		}
		sum := getSum(fpath, h)
		fpath = strings.TrimPrefix(fpath, dir+release+"/")

		mux.Lock()
		hashMap[fpath] = sum
		mux.Unlock()
	}
}
