package spellcheck

import (
	"sync"
	"strconv"
)

var Mu sync.RWMutex

type WordItem struct {
	Word string 
	Prob float64
}

type Response struct {
	Word string
	Message string
}

type WordList []*WordItem

func (wl WordList) Len () int {
	return len(wl)
}

func (wl WordList) Swap (x, y int) {
	wl[x], wl[y] = wl[y], wl[x]
}

func (wl WordList) Less (x, y int) bool {
	return wl[x].Word < wl[y].Word
}

var GlobalWordList = WordList{}


func CorrectWord (input, prefix string, debug bool) *Response {
	if len(input) < 1 {
		return &Response{input, ""}
	}
	
	// Convert uppercase to lowercase
	bTmp := make([]byte,0)
	for i := 0; i < len(input); i++ {
		if input[i] > 64 && input[i] < 91 {
			bTmp = append(bTmp, input[i] + 32)
		} else {
			bTmp = append(bTmp, input[i])
		}
	}

	input = string(bTmp)

	firstByte := input[0]

	lowRange := len(input) - 2
	if lowRange < 0 {
		lowRange = 0
	}

	highRange := len(input) + 2
  	
  	// Create local copy of word list 
  	Mu.RLock()
  	localWordList := make([]*WordItem, len(GlobalWordList))
  	copy(localWordList, GlobalWordList)
  	Mu.RUnlock()

  	// Store index of best 1st choice.
  	max := 0.0
	maxIdx := 0

	// Store index of best 2nd choice.
	secondMax := 0.0
	secondMaxIdx := 0

	// Store index of best 3rd choice.
	thirdMax := 0.0
	thirdMaxIdx := 0

  	for i := 0; i < len(localWordList); i++ {
  		word := localWordList[i].Word
  		prob := localWordList[i].Prob
  		length := len(word)
  		
  		if length < 1 || word[0] != firstByte {
  			continue
  		}
  		
  		if length > lowRange && length < highRange {
  			check := oneByteDifference(input, word)
			if check {
				max = prob + 0.0001
				maxIdx = i
				continue
			} else if prob > thirdMax {
				thirdMax = prob 
  				thirdMaxIdx = i
  				continue	
			}
  		}
  		
  		if len(word) >= len(prefix) &&  word[:len(prefix)] == prefix && prob > secondMax {
  			secondMax = prob 
  			secondMaxIdx = i	
  		}
  	}
	
	// use highest prob where found word is just one character difference from input
	if max > 0 {
		msg := ""
		if debug {
			msg = "Debug Message: Solved by manipulating one character. Prob: " + strconv.FormatFloat(max*100.00, 'g', 6, 64) + "%"
		}
		return &Response{localWordList[maxIdx].Word, msg}
	}

	// use highest prob where entire prefix matches
	if secondMax > 0 {
		msg := ""
		if debug {
			msg = "Debug Message: Solved using prefix: '" + prefix + "'. Prob: " + strconv.FormatFloat(secondMax*100.00, 'g', 6, 64) + "%"
		}
		return &Response{localWordList[secondMaxIdx].Word, msg}
	} 

	// last resort just use highest prob from words that start with same letter.
	if thirdMax > 0 {
		msg := ""
		if debug {
			msg = "Debug Message: Best guess using all" + string(firstByte) + "words. Prob:" + strconv.FormatFloat(thirdMax, 'g', 6, 64) + "%"
		}
		return &Response{localWordList[thirdMaxIdx].Word, msg}
	}

	return nil
}

func transposeSwap (tmpB []byte, a, b int) {
	tmpB[a], tmpB[b] = tmpB[b], tmpB[a] 
} 

// Manipulate input by one byte, either delete, transpose, replace, insert, or append.
// Check if manipulated input matches word.
func oneByteDifference (input, word string) bool {
	
	// Prebuild slice of bytes representing lowercase alphabet, i.e. alphabytes.
	alphaBytes := make([]byte, 0)
	for i := 97; i < 123; i++ {
		alphaBytes = append(alphaBytes, byte(i))
	}

	for i := 1; i < len(input); i++ {
		// Delete
		delete := input[:i] + input[i+1:]

		if delete == word {
			return true
		}

		// Transpose
		if i > 1 {
			tranpose := make([]byte, len([]byte(input)))
			copy(tranpose, []byte(input)) 
			transposeSwap(tranpose, i-1, i)

			if string(tranpose) == word {
				return true
			}
		}

		// Replace or Insert
		for j := 0; j < len(alphaBytes); j++ {
			replacement := make([]byte, len([]byte(input)))
			copy(replacement, []byte(input))
			replacement[i] = alphaBytes[j]

			insert := input[:i] + string(alphaBytes[j]) + input[i:]

			if string(replacement) == word || insert == word {
				return true
			}
		}
	}

	// Append
	for j := 0; j < len(alphaBytes); j++ {
		addToEnd := input + string(alphaBytes[j])
		if addToEnd == word {
			return true
		}
	}

	return false
}