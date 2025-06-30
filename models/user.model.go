package userModel

import (
	"REST-API_go/db"
	"errors"

	"golang.org/x/crypto/bcrypt"
)


type User struct {
	ID       int
	Email string
	Password string
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateUser(username, password string) error {
	hashedPassword, err := HashPassword(password)
	if err != nil {
	_, err = db.DB.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", username, hashedPassword) // Ensure DB is initialized before calling this
	}

	_, err = db.DB.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", username, hashedPassword)
	return err
}

func GetUserByEmail(username string) (*User, error) {
	user := User{}
	err := db.DB.QueryRow("SELECT id, username, password FROM users WHERE username=$1", username).
		Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}
