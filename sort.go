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

	if rel1[2] == "GA" {
		return false
	}

	if rel2[2] == "GA" {
		return true
	}

	if 3 < len(rel1) {
		if len(rel1) == len(rel2) { // RC, Snap
			if rel1[2] == "RC" {
				if rel2[2] == "RC" {
					rcVer1, err := strconv.Atoi(rel1[4])
					if err != nil {
						log.Fatal(err)
					}

					rcVer2, err := strconv.Atoi(rel2[4])
					if err != nil {
						log.Fatal(err)
					}

					if rcVer1 < rcVer2 {
						return true
					}
				}
			} else if rel2[2] == "RC" {
				return true
			} else { // both Snaps
				snapVer1, err := strconv.Atoi(rel1[4])
				if err != nil {
					log.Fatal(err)
				}

				snapVer2, err := strconv.Atoi(rel2[4])
				if err != nil {
					log.Fatal(err)
				}

				if snapVer1 < snapVer2 {
					return true
				}
			}
		} else { // rel1 is RC or Snap, rel2 is either Alpha or Beta
			return false
		}
	} else if 3 < len(rel2) { // rel1 is Alpha or Beta, rel2 is RC or Snap
		return true
	}
	return false
}
