package main

import (
	"os"
	"fmt"
	"time"
	"bufio"
	"context"
	"os/signal"
	
	"go_dictionary/dictionary"

	Z "github.com/goprammer/Z"
)

func handleStdin () {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Please enter a search word:\n")
		searchTerm, _ := r.ReadString('\n')
		searchTerm = searchTerm[:len(searchTerm) - 1]
		idx := Z.FirstIndex(" -d", searchTerm)
		if idx == -1 {
			idx = Z.FirstIndex(" --debug", searchTerm)
		}

		debug := false
		if idx != -1 {
			searchTerm = searchTerm[:idx]
			debug = true
		}

		ansCh := make(chan *dictionary.Answer)
		altCh := make(chan *dictionary.Answer)
		errCh := make(chan error)

		go dictionary.T.Search(ansCh, altCh, errCh, []byte(searchTerm), debug)
		
		select {
		case answer := <- ansCh:
			answer.Print()
		case answer := <- altCh:
			answer.Print()
		case err := <- errCh:
			fmt.Println(err)
		}
		fmt.Println()
	}
}

func main () {
	fmt.Println(dictionary.T.WordCount())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	
	go handleStdin()
	
	select {
	case <- ctx.Done():
		fmt.Printf("\nGracefully shutting down...\n")
		time.Sleep(3 * time.Second)
	}
}
