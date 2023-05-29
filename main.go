package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/mar-cial/items/handler"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("here", err)
	}
}

func main() {
	app := handler.NewApp()
	router := handler.CreateRouter(app)

	server := http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	fmt.Println("Running on port 8000")
	log.Fatalln(server.ListenAndServe())
}
