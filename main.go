// main.go
package main

import (
	dataBase "REST-API_go/db"
	"REST-API_go/handlers"
	config "REST-API_go/minio-config"

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

	// Initialize MinIO client
	config.InitMinio()

	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.Handle("/protected", auth.AuthMiddleware(http.HandlerFunc(handlers.ProtectedHandler)))

	http.HandleFunc("/profile/upload", handlers.UploadHandler)
	http.HandleFunc("/profile/download", handlers.DownloadHandler)
	http.HandleFunc("/profile/update", handlers.UpdateHandler)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
