package main

import (
	"couponcutter/authetication"
	"couponcutter/listing"
	"couponcutter/rest"
	"couponcutter/storage/database"
	"couponcutter/storemanagement"
	"fmt"
	"log"
	"os"

	"github.com/go-chi/httplog"
	"github.com/rs/zerolog"
)

const SECRET_KEY = "57574F2B696AE7C7A29ggggggggggggggggggggggdddddlmbd5B5CBF289324"

func main() {
	storage, err := database.NewStorage()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	listing := listing.NewService(storage)
	auth := authetication.NewService(storage, SECRET_KEY)
	sManager := storemanagement.NewService(storage)
	apiLogger := httplog.NewLogger("web-server", httplog.Options{
		Concise: true,
	})
	run(listing, sManager, auth, apiLogger, "127.0.0.1:5500")

}

func run(listing listing.Service,
	sManager storemanagement.Service,
	auth authetication.Service,
	logger zerolog.Logger,
	addr string) {

	server := rest.NewServer(logger, listing, sManager, auth, addr)
	fmt.Printf("starting server at http://%v \n", addr)
	log.Fatal(server.ListenAndServe())

}
