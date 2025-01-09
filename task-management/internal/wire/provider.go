package wire

import (
	core_grpcclient "github.com/cnc-csku/task-nexus/go-lib/grpcclient"
	"github.com/cnc-csku/task-nexus/task-management/config"
	"github.com/cnc-csku/task-nexus/task-management/domain/services"
	"github.com/cnc-csku/task-nexus/task-management/internal/adapters/repositories/grpcclient"
	"github.com/cnc-csku/task-nexus/task-management/internal/adapters/repositories/mongo"
	"github.com/cnc-csku/task-nexus/task-management/internal/adapters/repositories/storage"
	"github.com/cnc-csku/task-nexus/task-management/internal/adapters/rest"
	"github.com/cnc-csku/task-nexus/task-management/internal/infrastructure/database"
	"github.com/cnc-csku/task-nexus/task-management/internal/infrastructure/llm"
	"github.com/cnc-csku/task-nexus/task-management/internal/infrastructure/router"
	core_storage "github.com/cnc-csku/task-nexus/task-management/internal/infrastructure/storage"
	"github.com/google/wire"
)

var CtxSet = wire.NewSet(
	NewCtx,
)

var ConfigSet = wire.NewSet(
	config.NewConfig,
)

var InfraSet = wire.NewSet(
	database.NewMongoClient,
	router.NewRouter,
	llm.NewOllamaClient,
	core_storage.NewMinIOClient,
)

var RepositorySet = wire.NewSet(
	mongo.NewMemberRepository,
	storage.NewMinioRepository,
)

var ServiceSet = wire.NewSet(
	services.NewCommonService,
	services.NewMemberService,
)

var RestHandlerSet = wire.NewSet(
	rest.NewHealthCheckHandler,
	rest.NewCommonHandler,
	rest.NewMemberHandler,
)

var GrpcClientSet = wire.NewSet(
	config.ProvideGrpcClientConfig,
	core_grpcclient.NewGrpcClient,
	grpcclient.NewGrpcClient,
)
