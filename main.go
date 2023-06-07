package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mar-cial/items/api"
)

func main() {
	srv, err := api.CreateServer()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Serving on port %s\n", os.Getenv("SERVERPORT"))
	log.Fatalln(srv.ListenAndServe())
}
