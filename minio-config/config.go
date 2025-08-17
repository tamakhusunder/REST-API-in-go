package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client
var BucketName = "profile"

func InitMinio() {

	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := false

	// Initialize minio client object.
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		log.Fatalln("Error initialize minio client: ", err)
	}

	fmt.Println("Connected to Minio 🎉")
	fmt.Printf(accessKeyID, secretAccessKey, endpoint)

	MinioClient = client
	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, BucketName)
	if err != nil {
		log.Fatalf("❌ Could not check bucket: %v", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, BucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("❌ Could not create bucket: %v", err)
		}
		log.Printf("✅ Created bucket: %s\n", BucketName)
	} else {
		log.Printf("ℹ️ Using existing bucket: %s\n", BucketName)
	}
}
