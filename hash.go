package main

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var mux sync.Mutex

func hashRelease(dir, release string, jobs int) (map[string]string, error) {
	var wg sync.WaitGroup
	wq := make(chan string, jobs+1)
	hashedRelease := make(map[string]string)

	go getFiles(dir+release+"/", wq)

	for workers := 0; workers < jobs; workers++ {
		wg.Add(1)
		go hashFile(hashedRelease, wq, &wg)
	}

	wg.Wait()

	return hashedRelease, nil
}

func getFiles(dir string, wq chan<- string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			getFiles(dir+file.Name()+"/", wq)
		} else {
			wq <- (dir + file.Name())
		}
	}

	close(wq)
}

func hashFile(hashMap map[string]string, wq <-chan string, wg *sync.WaitGroup) {
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

		sum := getSum(fpath, h)

		mux.Lock()
		hashMap[fpath] = sum
		mux.Unlock()
	}
}
