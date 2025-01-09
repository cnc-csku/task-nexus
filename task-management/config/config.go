package config

import (
	"log"

	"github.com/caarlos0/env/v11"
	coreGrpcClient "github.com/cnc-csku/task-nexus/go-lib/grpcclient"
	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName  string                          `env:"SERVICE_NAME"`
	RestServer   RestServerConfig                `envPrefix:"REST_SERVER_"`
	MongoDB      MongoDBConfig                   `envPrefix:"MONGO_"`
	GrpcServer   GrpcServerConfig                `envPrefix:"GRPC_SERVER_"`
	GrpcClient   coreGrpcClient.GrpcClientConfig `envPrefix:"GRPC_CLIENT_"`
	OllamaClient OllamaClientConfig              `envPrefix:"OLLAMA_CLIENT_"`
	MinioClient  MinioClientConfig               `envPrefix:"MINIO_"`
}

type RestServerConfig struct {
	Port string `env:"PORT"`
}

type MongoDBConfig struct {
	URI      string `env:"URI"`
	Port     string `env:"PORT"`
	Username string `env:"USERNAME"`
	Password string `env:"PASSWORD"`
	Database string `env:"DATABASE"`
}

type GrpcServerConfig struct {
	Port           string `env:"PORT"`
	MaxSendMsgSize int    `env:"MAX_SEND_MSG_SIZE"`
	MaxRecvMsgSize int    `env:"MAX_RECV_MSG_SIZE"`
	UseReflection  bool   `env:"USE_REFLECTION"`
}

type OllamaClientConfig struct {
	Endpoint      string `env:"ENDPOINT"`
	UseProxy      bool   `env:"USE_PROXY"`
	HttpProxyHost string `env:"HTTP_PROXY_HOST"`
	HttpProxyPort string `env:"HTTP_PROXY_PORT"`
}

type MinioClientConfig struct {
	Endpoint              string `env:"ENDPOINT"`
	AccessKeyID           string `env:"ACCESS_KEY_ID"`
	SecretAccessKey       string `env:"SECRET_ACCESS_KEY"`
	BucketName            string `env:"BUCKET_NAME"`
	UseSSL                bool   `env:"USE_SSL"`
	FileUploadSizeLimitMB int64  `env:"FILE_UPLOAD_SIZE_LIMIT_MB"`
	PresignedURLExpirySec int64  `env:"PRESIGNED_URL_EXPIRY_SEC" envDefault:"900"`
	UseProxy              bool   `env:"USE_PROXY"`
	ProxyHost             string `env:"PROXY_HOST"`
	ProxyPort             string `env:"PROXY_PORT"`
}

func NewConfig() *Config {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it. Falling back to system environment variables.")
	}

	config := &Config{}

	// Parse environment variables into the config struct
	if err := env.Parse(config); err != nil {
		log.Fatalln("Failed to parse environment variables into Config struct:", err)
	}

	log.Printf("Config: %+v\n", config)
	return config
}

func ProvideGrpcClientConfig(config *Config) coreGrpcClient.GrpcClientConfig {
	return config.GrpcClient
}
