package main

import (
	"os"
	"log"
	"time"
	"context"
	"syscall"
	"net/http"
	"os/signal"
	"encoding/json"

	"go_dictionary/dictionary"
	
	"github.com/gorilla/mux"
	"github.com/go-openapi/runtime/middleware"
)

var PublicDomain string // Domain and port public uses to reach this container.

type Incoming struct {
	SearchTerm string `json: "searchTerm"`
	Debug bool `json: "debug"`
}

func Lookup (w http.ResponseWriter, r *http.Request) {
	incoming := &Incoming{}
	err := json.NewDecoder(r.Body).Decode(incoming)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		json.NewEncoder(w).Encode("You've send an empty string. Please visit " + PublicDomain + "/docs")
		return
	}

	ansCh := make(chan *dictionary.Answer)
	altCh := make(chan *dictionary.Answer)
	errCh := make(chan error)

	go dictionary.T.Search(ansCh, altCh, errCh, []byte(incoming.SearchTerm), incoming.Debug)
	
	select {
	case answer := <- ansCh:
		json.NewEncoder(w).Encode(answer) // 200
	case answer := <- altCh:
		json.NewEncoder(w).Encode(answer) // 200
	case err := <- errCh:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err.Error()) // 400
	}
}

func main () {
	PublicDomain = os.Getenv("PUBLIC_DOMAIN") 
	podPort := os.Getenv("DICTIONARY_API_PORT")
	mux := mux.NewRouter()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	server := http.Server{
		Handler: mux,
		Addr: podPort,
		ReadTimeout: 3 * time.Second,
		WriteTimeout: 3 * time.Second,
		IdleTimeout: 10 * time.Second,
	}

	opts := middleware.RedocOpts{SpecURL: "/swagger.yaml"}

	getMux := mux.Methods(http.MethodGet).Subrouter()
	getMux.HandleFunc("/v1.0", Lookup)
	getMux.Handle("/docs", middleware.Redoc(opts, nil))
	getMux.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	go func () {
		log.Println("Dictionary is listening on port: ", podPort)
		log.Fatal(server.ListenAndServe())	
	}()

	select {
	case <- ctx.Done():
		log.Printf("Gracefully shutting down...\n")
		time.Sleep(3 * time.Second)
	}
}