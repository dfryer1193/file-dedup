package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type hashStore struct {
	mux     sync.RWMutex
	fileMap map[string]string
}

var releaseRegex = `(\.[0-9])+-(GA|RC|(Internal)?Snapshot|Beta|Alpha)(-[0-9](\.[0-9])*)?$`

func getReleases(dir string, rel int) ([]string, error) {
	var releases []string
	var regexStr string

	if rel != 0 {
		regexStr = `^` + strconv.Itoa(rel) + releaseRegex
	} else {
		regexStr = `^[0-9]+` + regexStr
	}

	r := regexp.MustCompile(regexStr)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		matched := r.MatchString(f.Name())

		if matched {
			release := f.Name()
			releases = append(releases, release)
		}
	}

	return releases, nil
}

func getFiles(dir string, releases []string, ret chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, rel := range releases {
		searchFiles(dir+rel+"/", ret)
	}
	close(ret)
}

func searchFiles(dir string, ret chan<- string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			searchFiles(dir+file.Name()+"/", ret)
		} else {
			ret <- (dir + file.Name())
		}
	}
}

func dedup(base, dup string, silent, dryRun bool) {
	baseInfo, err := os.Stat(base)
	if err != nil {
		log.Fatal(err)
	}

	dupInfo, err := os.Stat(dup)
	if err != nil {
		log.Fatal(err)
	}

	if os.SameFile(baseInfo, dupInfo) {
		if silent {
			return
		}
		fmt.Printf("Skipping identical file %s\n", dupInfo.Name())
		return
	}

	fmt.Printf("Deduplicating %s...\n", dupInfo.Name())

	err = os.Rename(dup, dup+".bak")
	if err != nil {
		fmt.Printf("Could not rename %s, skipping...\n", dup)
		return
	}

	err = os.Link(base, dup)
	if err != nil {
		fmt.Printf("Failed to link file %s to %s\n", base, dup)
		err2 := os.Rename(dup+".bak", dup)
		if err2 != nil {
			fmt.Printf("Could not restore %s!\n", dup)
		}
		return
	}

	err = os.Remove(dup + ".bak")
	if err != nil {
		fmt.Printf("Failed to remove backup files %s.bak\n", dup)
	}

	return
}

func consumeFile(silent, dryRun bool, files *hashStore, wq <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	getSum := func(path string) string {
		h := sha256.New()
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
		sum := getSum(fpath)

		files.mux.RLock()
		fname := files.fileMap[sum]
		files.mux.RUnlock()

		if fname == "" {
			files.mux.Lock()
			files.fileMap[sum] = fpath
			files.mux.Unlock()
		} else {
			dedup(fname, fpath, silent, dryRun)
		}
	}
}

func dedupDir(dir string, rel, jobs int, silent, dryRun bool) {
	var wg sync.WaitGroup
	files := new(hashStore)
	wq := make(chan string, jobs+1)
	files.fileMap = make(map[string]string)

	releases, err := getReleases(dir, rel)
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	go getFiles(dir, releases, wq, &wg)

	for workers := 1; workers < jobs; workers++ {
		wg.Add(1)
		go consumeFile(silent, dryRun, files, wq, &wg)
	}

	wg.Wait()
}

func main() {
	var directory string
	var release, jobs int
	var silent, dryRun bool

	flag.StringVar(&directory, "dir", "./", "The directory containing releases to dedup.")
	flag.IntVar(&release, "rel", 0, "The release to dedup. 0 for all releases.")
	flag.IntVar(&jobs, "j", runtime.NumCPU(), "Maximum number of jobs.")
	flag.BoolVar(&silent, "s", false, "Run silently")
	flag.BoolVar(&dryRun, "d", false, "Do not perform deduplication.")
	flag.Parse()

	if !strings.HasSuffix(directory, "/") {
		directory = directory + "/"
	}

	dedupDir(directory, release, jobs, silent, dryRun)
}
