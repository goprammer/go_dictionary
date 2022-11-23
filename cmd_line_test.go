package main 

import (
	"time"
	"testing"
	"dictionary_API/dictionary"
)

/* Test cmd_line.go */
func Test_Dictionary_Match (t *testing.T) {
	searchTerms := []string{"superlative", "superlitive", "dog", "dogz", "superduper", "ice planet"}
	start := time.Now().UTC()
	for _,e := range searchTerms {
		ansCh := make(chan *dictionary.Answer)
		altCh := make(chan *dictionary.Answer)
		errCh := make(chan error)

		go dictionary.T.Search(ansCh, altCh, errCh, []byte(e), true)
		select {
		case answer := <- ansCh:
			t.Log(time.Now().UTC().Sub(start), "Match:", answer.Message)
		case answer := <- altCh:
			t.Log(time.Now().UTC().Sub(start), "Corrected Match:", answer.Message)
		case err := <- errCh:
			t.Error(time.Now().UTC().Sub(start), err)
		}

	}
	
	t.Log("Consider the Pass time, not the total time, because total time includes the load time for the csv dictionary and frequency map.")
}