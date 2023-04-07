package main

import (
	"fmt"
	"os"
	"sort"
)

const (
	HEIGHT = 10
	WIDTH  = 30
)

type code string
type letter string

var (
	alphabet = make(map[code]letter)
	morse    = make(map[letter]code)
)

func init() {
	alphabet[".-"] = "a"
	alphabet["-..."] = "b"
	alphabet["-.-."] = "c"
	alphabet["-.."] = "d"
	alphabet["."] = "e"
	alphabet["..-."] = "f"
	alphabet["--."] = "g"
	alphabet["...."] = "h"
	alphabet[".."] = "i"
	alphabet[".---"] = "j"
	alphabet["-.-"] = "k"
	alphabet[".-.."] = "l"
	alphabet["--"] = "m"
	alphabet["-."] = "n"
	alphabet["---"] = "o"
	alphabet[".--."] = "p"
	alphabet["--.-"] = "q"
	alphabet[".-."] = "r"
	alphabet["..."] = "s"
	alphabet["-"] = "t"
	alphabet["..-"] = "u"
	alphabet["...-"] = "v"
	alphabet[".--"] = "w"
	alphabet["-..-"] = "x"
	alphabet["-.--"] = "y"
	alphabet["--.."] = "z"

	alphabet["-----"] = "0"
	alphabet[".----"] = "1"
	alphabet["..---"] = "2"
	alphabet["...--"] = "3"
	alphabet["....-"] = "4"
	alphabet["....."] = "5"
	alphabet["-...."] = "6"
	alphabet["--..."] = "7"
	alphabet["---.."] = "8"
	alphabet["----."] = "9"

	alphabet["-...-"] = "="
	alphabet["-..-."] = "/"
	alphabet["..--.."] = "?"
	alphabet["--..--"] = ","
	alphabet[".-.-.-"] = "."
	alphabet["-.--."] = "("
	alphabet["-.--.-"] = ")"
	alphabet[".-..."] = "&"
	alphabet[".-.-."] = "+"
	alphabet[".--.-."] = "@"
	alphabet["---..."] = ":"
	alphabet["-.-.-."] = ";"
	// Prosigns:
	alphabet[".-.-"] = "<aa>"
	alphabet["........"] = "<hh>"

	for code, letter := range alphabet {
		morse[letter] = code
	}
}

type Element struct {
	d int
	s bool
}

func detectCode(ds []Element) string {
	ditMean, dahMean := classifySignals(ds)
	ditGap := ditMean
	charGap := dahMean
	wordGap := 7 * ditGap
	res := make([]byte, 0, len(ds))
	char := make([]byte, 0, 4)
	str := ""
	for i := 0; i < len(ds); i++ {
		if ds[i].s {
			if abs(ds[i].d-ditMean) < abs(ds[i].d-dahMean) {
				res = append(res, '.')
				char = append(char, '.')
			} else {
				res = append(res, '-')
				char = append(char, '-')
			}
		} else {
			if abs(ds[i].d-charGap) < abs(ds[i].d-ditGap) && abs(ds[i].d-charGap) < abs(ds[i].d-wordGap) {
				res = append(res, '>')
				str += string(alphabet[code(char)])
				char = char[0:0]
			}
			if abs(ds[i].d-wordGap) < abs(ds[i].d-ditGap) && abs(ds[i].d-wordGap) < abs(ds[i].d-charGap) {
				res = append(res, '<')
				str += string(alphabet[code(char)])
				str += " "
				char = char[0:0]
			}
		}
	}
	str += string(alphabet[code(char)])
	return str
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func classifyGaps(ds []Element) (ditGap int, charGap int, wordGap int) {
	f, err := os.Create("classifyGaps.log")
	if err != nil {
		os.Exit(1)
	}
	defer f.Close()
	unsortedGaps := make([]int, 0)
	gaps := make([]int, 0)
	for _, d := range ds {
		if !d.s {
			unsortedGaps = append(unsortedGaps, int(d.d))
			gaps = append(gaps, int(d.d))
		}
	}
	sort.IntSlice(gaps).Sort()

	lastDitGap := gaps[0]
	lastCharGap := lastDitGap * 3
	lastWordGap := gaps[len(gaps)-1]

	fmt.Fprintf(f, "Gaps: %v\n", gaps)

	for {
		fmt.Fprintf(f, "DitGap: %v, CharGap: %v, WordGap: %v\n", lastDitGap, lastCharGap, lastWordGap)
		border1 := 0
		border2 := 0
		for i, s := range gaps {
			if abs(s-lastDitGap) > abs(s-lastCharGap) {
				border1 = i
				break
			}
		}
		for i, s := range gaps {
			if abs(s-lastCharGap) > abs(s-lastWordGap) {
				border2 = i
				break
			}
		}
		fmt.Fprintf(f, "border1: %v, border2: %v\n", border1, border2)
		ditGapMean := 0
		for i := 0; i < border1; i++ {
			ditGapMean += gaps[i]
		}
		ditGapMean /= border1

		charGapMean := 0
		for i := border1; i < border2; i++ {
			charGapMean += gaps[i]
		}
		charGapMean /= border2 - border1

		wordGapMean := 0
		for i := border2; i < len(gaps); i++ {
			wordGapMean += gaps[i]
		}
		wordGapMean /= len(gaps) - border2

		if ditGapMean == lastDitGap && charGapMean == lastCharGap && wordGapMean == lastWordGap {
			break
		}
		lastDitGap = ditGapMean
		lastCharGap = charGapMean
		lastWordGap = wordGapMean
	}

	return lastDitGap, lastCharGap, lastWordGap
}

// K-means for classifying dots and dashes
func classifySignals(ds []Element) (int, int) {
	unsortedSignals := make([]int, 0)
	signals := make([]int, 0)
	for _, d := range ds {
		if d.s {
			unsortedSignals = append(unsortedSignals, int(d.d))
			signals = append(signals, int(d.d))
		}
	}
	sort.IntSlice(signals).Sort()

	lastDotMean := signals[0]
	lastDahMean := signals[len(signals)-1]
	for {
		border := 0
		for i, s := range signals {
			if abs(s-lastDotMean) > abs(s-lastDahMean) {
				border = i
				break
			}
		}
		ditMean := 0
		for i := 0; i < border; i++ {
			ditMean += signals[i]
		}
		ditMean /= border
		dahMean := 0
		for i := border; i < len(signals); i++ {
			dahMean += signals[i]
		}
		dahMean /= len(signals) - border
		if ditMean == lastDotMean && dahMean == lastDahMean {
			break
		}
		lastDotMean = ditMean
		lastDahMean = dahMean
	}
	return lastDotMean, lastDahMean
}
