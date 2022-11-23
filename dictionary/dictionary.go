package dictionary

// Select waits for match, best guess, or error.
// Handle erroneous client input.
// Work is done by changing values of pointers to structs.
// Graceful shutdown with context.
// Nexting and error channel send errors to front.
// Reading system files with tricky formatting.

// build a similated network of ping times and routing, then asynchronously query them, feed a graph and find min spanning tree

import (
	"os"
	"fmt"
	"sync"
	"sort"
	"bufio"
	"errors"
	"strconv"
	
	sp "go_dictionary/spellcheck"
)

// Trie holds 39 characters: a-z, 0-9, -, ', \s between words. 
const (
	N = 39
)

type Node struct {
	Key [N]*Node
	End bool
	Definitions []string
}

type Trie struct {
	Root *Node
	Count int
	Mu sync.RWMutex
}

type SearchResults struct {
	Index int
	Node *Node
}

type Answer struct {
	Message string
	Definitions []string
}

var T Trie


func init () {
	// Build a frequency map for words in a book.
	FrequencyMap := make(map[string]int)
	d := os.Getenv("PWD") + "/Gutenburg_EBook"
	f, err := os.Open(d + "/TheAdventuresOfSherlockHolmes.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	r := bufio.NewReader(f)
	
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		
		// Allow lowercase letters, numbers, whitespace and newline into frequency map.
		bWord := make([]byte, 0)
		for i := 0; i < len(line); i++ {
			b := line[i]
			switch {
			case b > 96 && b < 123:
				bWord = append(bWord, b)
			case b > 64 && b < 91:	
				bWord = append(bWord, b+32)
			case b == 39 || b == 45:
				bWord = append(bWord, b)
			case b == 32 || b == 10:
				if len(bWord) > 0 {
					FrequencyMap[string(bWord)]++	
				}
				bWord = make([]byte, 0)
			}
		}	
	}

	T.Root = &Node{}
	T.Count = 0
	
	// Dictionary from csv format to trie.
	directory := os.Getenv("PWD") + "/Gutenburg_Dictionary"
	files, err := os.ReadDir(directory)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	wordMap := make(map[string]bool)

	for _,file := range files {
		f, err := os.Open(directory + "/" + file.Name())
		if err != nil {
			fmt.Println(err)
			return
		}
		r := bufio.NewReader(f)
		
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}

			if len(line) < 3 {
				continue
			}

			switch {
			case line[0] == 34 && line[len(line)-3] == 34:
				line = line[1:len(line)-3]
			case line[0] == 34:
				line = line[1:len(line)-2] 
			default:
				line = line[:len(line)-2]
			}
			
			
			word := ""
			definition := ""
			for i,e := range line {
				if e == 40 {
					word = line[:i-1]
					definition = line[i:]
					break
				}
			}
			if len(word) == 0 {
				continue
			}
			
			T.Add([]byte(word), definition)

			bTmp := make([]byte,0)
			for i := 0; i < len(word); i++ {
				if word[i] > 64 && word[i] < 91 {
					bTmp = append(bTmp, word[i] + 32)
				} else {
					bTmp = append(bTmp, word[i])
				}
			}
			word = string(bTmp)
			
			_, ok := wordMap[word]
			if ok {
				continue
			}

			wordMap[word] = true
			prob := 0.000001
			f, ok := FrequencyMap[word]
			if ok {
				prob = float64(f)/float64(len(FrequencyMap))	
			}

			sp.GlobalWordList = append(sp.GlobalWordList, &sp.WordItem{word, prob})
		}
		
	}
	sort.Sort(sp.GlobalWordList)
}


func (t *Trie) Add (word []byte, definition string) {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	node := t.Root

	for _, e := range word {
		e, err := ASCII_To_Trie(e)
		if err != nil {
			return 
		}
		if node.Key[e] == nil {
			node.Key[e] = &Node{}
		}
		node = node.Key[e]
	}

	if !node.End {
		t.Count++	
		node.End = true
	}
	
	node.Definitions = append(node.Definitions, definition)
}


func (t *Trie) Search (ansCh, altCh chan<- *Answer, errCh chan<- error, word []byte, debug bool) {
	defer close(ansCh)
	defer close(altCh)
	defer close(errCh)

	t.Mu.RLock()
	defer t.Mu.RUnlock()
	node := t.Root
	
	word = trimWhiteSpaceFrontBack(word)

	// Use suupplied word to search for definition. 
	searchResults, err := travTrie(word, node)
	if err != nil {
		errCh <- err
		return
	}
	
	// Found definition.
	// Return it through answer channel.
	if searchResults.Node.End && searchResults.Index == len(word) - 1 {
		ansCh <- &Answer{string(word) + ":\n", searchResults.Node.Definitions}
		return
	}

	// Couldn't find definition.
	// Create prefix from correctly matched characters.
	prefix := ""
	if len(word) >= searchResults.Index + 1 {
		prefix = string(word[:searchResults.Index+1])
	} else {
		prefix = string(word)
	}
	
	// Found most probably correct word
	res := sp.CorrectWord(string(word), prefix, debug)

	// Retrieve definition with correct word.
	searchResults, err = travTrie([]byte(res.Word), t.Root)
	if err != nil {
		errCh <- err
		return
	}
	
	// Return definition through alternate channel with corrected word and explanation.
	if searchResults.Node.End {
		res.Message = res.Message + "\nCoundn't find '" + string(word) + "'\nUsing '" + res.Word + "' instead."
		altCh <- &Answer{res.Message, searchResults.Node.Definitions}
		return
	} else {
		errCh <- err
		return	
	}	
}


func (t *Trie) WordCount () string {
	str := strconv.Itoa(t.Count)
	for i := len(str) - 3; i > 0; i = i - 3 {
		str = str[:i] + "," + str[i:]	
	}

	return str + " words in dictionary\n---"
}


func (answer *Answer) Print () {
	fmt.Printf("---\n%s", answer.Message)
		for _, e := range answer.Definitions {
			fmt.Println(e)	
		}
}


func ASCII_To_Trie (e byte) (byte, error) {
	e = lowercase(e)

	switch {
	case e > 96 && e < 123:
		return byte(e - 97), nil
	case e == 45:
	 	// hyphen
		return byte(26), nil 
	case e == 39:
		// apostrophy
		return byte(27), nil
	case e == 32:
		// middle whitespace
		return byte(28), nil	
	case e > 47 && e < 58:
		// digit
		return byte(e-19), nil		
	default:
		return byte(100), errors.New("Error: Dictionary can't handle this character: " + string(e))
	}
}


func travTrie (word []byte, node *Node) (*SearchResults, error) {
	i := 0
	for i < len(word) {
		norm, err := ASCII_To_Trie(word[i])
		if err != nil {
			return nil, err
		}
		
		if node.Key[norm] != nil {
			
			node = node.Key[norm]
		} else {
			i--
			break
		}

		if i != len(word) - 1 {
			i++
		} else {
			break
		}
	}

	return &SearchResults{i, node}, nil
}


func lowercase (e byte) byte {
	if e > 64 && e < 91 {
		return e + 32
	}
	return e
}

func trimWhiteSpaceFrontBack (input []byte) []byte {
	for len(input) > 0 && input[0] == 32 {
		input = input[1:]
	}

	for len(input) > 0 && input[len(input)-1] == 32 {
		input = input[:len(input)-1]
	}

	return input
}