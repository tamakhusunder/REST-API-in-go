// handlers.go
package handlers

import (
	// "REST-API_go/handlers"
	config "REST-API_go/minio-config"
	userModel "REST-API_go/models"
	"REST-API_go/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

var creds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

const MaxUploadSize = 500 * 1024 * 1024 // 500 MB

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

// POST /upload
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST", http.StatusMethodNotAllowed)
		return
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "❌ Upload error: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Reject if file size > 500MB
	if header.Size > MaxUploadSize {
		http.Error(w, "❌ File too large (max 500MB)", http.StatusRequestEntityTooLarge) // 413 = Payload Too Large
		return
	}

	info, err := config.MinioClient.PutObject(
		r.Context(),
		config.BucketName,
		header.Filename,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: header.Header.Get("Content-Type")},
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	reqParams := make(url.Values)
	// url, _ := config.MinioClient.PresignedGetObject(r.Context(), config.BucketName, header.Filename, 24*time.Hour, nil)

	presignedURL, err := config.MinioClient.PresignedGetObject(
		r.Context(),
		config.BucketName,
		header.Filename,
		24*time.Hour, // URL valid for 24 hours
		reqParams,
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, "✅ Uploaded %s (%d bytes)\n URL:%s \n", info.Key, info.Size, presignedURL)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"presignedUrl": presignedURL.String(),
	})
}

// GET /download?file=filename
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Use GET", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "Missing ?file=", 400)
		return
	}

	obj, err := config.MinioClient.GetObject(r.Context(), config.BucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(w, obj)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

// PUT /update?file=filename
// (re-uploads new content with same object name)
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Use PUT", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "Missing ?file=", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	_, err = config.MinioClient.PutObject(
		r.Context(),
		config.BucketName,
		filename,
		io.NopCloser(r.Body),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, "✅ Updated file %s\n", filename)
}
