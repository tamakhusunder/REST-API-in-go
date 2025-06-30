// handlers.go
package handlers

import (
	// "REST-API_go/handlers"
	userModel "REST-API_go/models"
	"REST-API_go/utils"
	"encoding/json"
	"net/http"
)

var creds struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	json.NewDecoder(r.Body).Decode(&creds)

	if creds.Email == "" || creds.Password == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	err := userModel.CreateUser(creds.Email, creds.Password)
	if err != nil {
		http.Error(w, "User exists or error occurred", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered"))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	json.NewDecoder(r.Body).Decode(&creds)

	user, err := userModel.GetUserByEmail(creds.Email)
	if err != nil || !userModel.CheckPasswordHash(user.Password, creds.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateJWT(user.Email)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	gmail := r.Context().Value("gmail").(string)
	w.Write([]byte("Welcome " + gmail))
}
