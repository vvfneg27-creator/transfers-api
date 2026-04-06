package main

import (
	"transfers-api/internal/clients"
	"transfers-api/internal/config"
	"transfers-api/internal/handlers"
	"transfers-api/internal/logging"
	"transfers-api/internal/repositories"
	"transfers-api/internal/services"
	"transfers-api/internal/transport"
	"transfers-api/internal/version"
)

func main() {
	// init logger
	logger := logging.Logger
	logger.Info("logger started")

	// init config
	cfg := config.ParseFromEnv()
	logger.Infof("config loaded: %v", cfg.String())

	// init repositories
	transfersDB := repositories.NewTransfersMongoDBRepository(cfg.MongoDBConfig)
	//transfersDB := repositories.NewTransfersMySqlDBRepository(cfg.MySQLDBConfig)
	//transfersCache := repositories.NewTransfersMemcachedRepository(cfg.MemcachedConfig)
	transfersCache := repositories.NewTransfersCcachedRepository(cfg.CcachedConfig)
	logger.Info("repositories created")

	// init clients
	tranfersDBPublisher := clients.NewRabbitMQClient(cfg.RabbitMQConfig)
	defer func() {
		if err := tranfersDBPublisher.Close(); err != nil {
			logger.Warnf("error closing RabbitMQ client: %v", err)
		}
	}()
	logger.Info("clients created")

	// init services
	transfersService := services.NewTransfersService(cfg.Business, transfersDB, transfersCache, tranfersDBPublisher)
	logger.Infof("services created")

	// init handlers
	transfersHandler := handlers.NewTransfersHandler(transfersService)
	logger.Infof("handlers created")

	// init server
	server := transport.NewHTTPServer(transfersHandler)
	server.MapRoutes()
	logger.Infof("server created, running %s@%s", version.AppName, version.Version)

	// run server
	server.Run(":8080")
}
