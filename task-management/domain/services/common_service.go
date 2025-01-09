package services

import (
	"github.com/cnc-csku/task-nexus/task-management/config"
	"github.com/cnc-csku/task-nexus/task-management/domain/repositories"
	"github.com/cnc-csku/task-nexus/task-management/internal/adapters/repositories/grpcclient"
)

type CommonService interface {
}

type commonService struct {
	config     *config.Config
	minioRepo  repositories.MinioRepository
	grpcClient *grpcclient.GrpcClient
}

func NewCommonService(
	config *config.Config,
	minioRepo repositories.MinioRepository,
	grpcClient *grpcclient.GrpcClient,
) CommonService {
	return &commonService{
		config:     config,
		minioRepo:  minioRepo,
		grpcClient: grpcClient,
	}
}
