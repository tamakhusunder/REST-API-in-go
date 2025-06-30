// main.go
package main

import (
	dataBase "REST-API_go/db"
	"REST-API_go/handlers"

	auth "REST-API_go/middlewares"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dataBase.InitDB()

	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.Handle("/protected", auth.AuthMiddleware(http.HandlerFunc(handlers.ProtectedHandler)))

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
