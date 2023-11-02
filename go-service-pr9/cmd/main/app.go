package main

import (
	"PR9/internal/adapters/api"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
		panic(err)
	}

	router := mux.NewRouter()

	http.Handle("/", router)
	api.SetupRoutes(router)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}
