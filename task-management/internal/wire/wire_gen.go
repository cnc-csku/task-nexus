// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package wire

import (
	"github.com/cnc-csku/task-nexus/task-management/config"
	"github.com/cnc-csku/task-nexus/task-management/domain/services"
	"github.com/cnc-csku/task-nexus/task-management/internal/adapters/repositories/mongo"
	"github.com/cnc-csku/task-nexus/task-management/internal/adapters/rest"
	"github.com/cnc-csku/task-nexus/task-management/internal/infrastructure/api"
	"github.com/cnc-csku/task-nexus/task-management/internal/infrastructure/database"
	"github.com/cnc-csku/task-nexus/task-management/internal/infrastructure/router"
)

// Injectors from wire.go:

func InitializeApp() *api.EchoAPI {
	context := NewCtx()
	configConfig := config.NewConfig()
	client := database.NewMongoClient(configConfig, context)
	healthCheckHandler := rest.NewHealthCheckHandler()
	routerRouter := router.NewRouter(healthCheckHandler)
	echoAPI := api.NewEchoAPI(context, configConfig, client, routerRouter)
	return echoAPI
}

func InitializeGrpcServer() *api.GrpcServer {
	context := NewCtx()
	configConfig := config.NewConfig()
	client := database.NewMongoClient(configConfig, context)
	memberRepository := mongo.NewMemberRepository(client)
	memberService := services.NewMemberService(memberRepository)
	grpcServer := api.NewGrpcServer(context, configConfig, memberService)
	return grpcServer
}