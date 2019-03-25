package main

import (
	"log"
	"strconv"
	"strings"
	"unicode"
)

type byRelease []string

func (b byRelease) Len() int      { return len(b) }
func (b byRelease) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byRelease) Less(i, j int) bool {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}

	rel1 := strings.FieldsFunc(b[i], f)
	rel2 := strings.FieldsFunc(b[j], f)

	maj1, err := strconv.Atoi(rel1[0])
	if err != nil {
		log.Fatal(err)
	}

	maj2, err := strconv.Atoi(rel2[0])
	if err != nil {
		log.Fatal(err)
	}

	if maj1 < maj2 {
		return true
	} else if maj2 < maj1 {
		return false
	}

	min1, err := strconv.Atoi(rel1[1])
	if err != nil {
		log.Fatal(err)
	}

	min2, err := strconv.Atoi(rel2[1])
	if err != nil {
		log.Fatal(err)
	}

	if min1 < min2 {
		return true
	} else if min2 < min1 {
		return false
	}

	// Both major and minor release are the same
	if strings.ToUpper(rel1[2]) == "GA" {
		return false
	}

	if strings.ToUpper(rel2[2]) == "GA" {
		return true
	}

	// Check for alpha.
	if strings.ToUpper(rel1[2]) == "ALPHA" {
		return true
	}

	if strings.ToUpper(rel2[2]) == "ALPHA" {
		return false
	}

	// Check for beta.
	if strings.ToUpper(rel1[2]) == "BETA" {
		return true
	}

	if strings.ToUpper(rel2[2]) == "BETA" {
		return false
	}

	// check for RC
	if strings.ToUpper(rel1[2]) == "RC" {
		if strings.ToUpper(rel2[2]) == "RC" {
			ver1, err := strconv.Atoi(rel1[3])
			if err != nil {
				log.Fatal(err)
			}
			ver2, err := strconv.Atoi(rel2[3])
			if err != nil {
				log.Fatal(err)
			}

			if ver1 < ver2 {
				return true
			}
		} else {
			return false
		}
	} else {
		if strings.ToUpper(rel2[2]) == "RC" {
			return true
		}

		ver1, err := strconv.Atoi(rel1[3])
		if err != nil {
			log.Fatal(err)
		}
		ver2, err := strconv.Atoi(rel2[3])
		if err != nil {
			log.Fatal(err)
		}

		if ver1 < ver2 {
			return true
		}

	}

	return false
}
