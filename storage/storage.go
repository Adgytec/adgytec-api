package storage

import (
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var SpaceStorage *minio.Client

func InitCloudStorage() (*minio.Client, error) {
	minioClient, err := minio.New(os.Getenv("SPACE_STORAGE_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("SPACE_STORAGE_ACCESS_KEY"), os.Getenv("SPACE_STORAGE_SECRET_KEY"), ""),
		Secure: true,
	})
	if err != nil {
		return nil, err
	}

	SpaceStorage = minioClient
	return minioClient, nil
}
