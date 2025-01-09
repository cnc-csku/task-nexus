package storage

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"github.com/cnc-csku/task-nexus/task-management/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinIOClient(ctx context.Context, cfg *config.Config) *minio.Client {
	var httpClient *http.Client

	if cfg.MinioClient.UseProxy {
		transport := &http.Transport{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http",
				Host:   cfg.MinioClient.ProxyHost + ":" + cfg.MinioClient.ProxyPort,
			}),
		}
		httpClient = &http.Client{
			Transport: transport,
		}
	} else {
		httpClient = &http.Client{
			Transport: http.DefaultTransport,
		}
	}

	client, err := minio.New(cfg.MinioClient.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.MinioClient.AccessKeyID, cfg.MinioClient.SecretAccessKey, ""),
		Secure:    cfg.MinioClient.UseSSL,
		Transport: httpClient.Transport,
	})
	if err != nil {
		panic(err)
	}

	// test connection
	exists, err := client.BucketExists(ctx, cfg.MinioClient.BucketName)
	if err != nil {
		log.Fatal("🚫 Cannot connect to MinIO | ", err)
	} else if !exists {
		log.Fatalf("🚫 Bucket %s does not exist", cfg.MinioClient.BucketName)
	} else {
		log.Println("✅ Connected to MinIO | Bucket:", cfg.MinioClient.BucketName)
	}

	return client
}
