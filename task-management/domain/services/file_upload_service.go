package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/cnc-csku/task-nexus/task-management/config"
	"github.com/cnc-csku/task-nexus/task-management/domain/repositories"
	"github.com/cnc-csku/task-nexus/task-management/domain/requests"
	"github.com/cnc-csku/task-nexus/task-management/domain/responses"
	"github.com/cnc-csku/task-nexus/task-management/internal/adapters/repositories/grpcclient"
	"github.com/google/uuid"
)

type FileUploadService interface {
	GeneratePutPresignedUrl(ctx context.Context, in *requests.GeneratePutPresignedUrlRequest) (*responses.GeneratePutPresignedUrlResponse, error)
}

type fileUploadService struct {
	config     *config.Config
	minioRepo  repositories.MinioRepository
	grpcClient *grpcclient.GrpcClient
}

func NewFileUploadService(
	config *config.Config,
	minioRepo repositories.MinioRepository,
	grpcClient *grpcclient.GrpcClient,
) FileUploadService {
	return &fileUploadService{
		config:     config,
		minioRepo:  minioRepo,
		grpcClient: grpcClient,
	}
}

func (s *fileUploadService) GeneratePutPresignedUrl(ctx context.Context, in *requests.GeneratePutPresignedUrlRequest) (*responses.GeneratePutPresignedUrlResponse, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found in context")
	}

	hashUUID, err := uuid.NewV7()
	if err != nil {

	}
	hash := sha256.Sum256([]byte(hashUUID.String()))

	var key string
	if in.IsPublic {
		key = fmt.Sprintf("public/%s/%s/%s", userId, hex.EncodeToString(hash[:]), in.Key)
	} else {
		key = fmt.Sprintf("%s/%s/%s", userId, hex.EncodeToString(hash[:]), in.Key)
	}

	expiredAt := time.Now().Add(time.Duration(s.config.MinioClient.PresignedURLExpirySec) * time.Second).Format(time.RFC3339)

	presignedURL, err := s.minioRepo.GetPutPresignedURL(ctx, key)
	if err != nil {
		return nil, err
	}

	return &responses.GeneratePutPresignedUrlResponse{
		Key:          key,
		PresignedUrl: presignedURL,
		ExpiredAt:    expiredAt,
	}, nil
}
