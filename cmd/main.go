package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/olad5/caution-companion/config"
	"github.com/olad5/caution-companion/config/data"
	loggingMiddleware "github.com/olad5/caution-companion/internal/handlers/logging"
	"github.com/olad5/caution-companion/internal/infra/cloudinary"
	"github.com/olad5/caution-companion/internal/infra/postgres"
	"github.com/olad5/caution-companion/internal/infra/redis"
	"github.com/olad5/caution-companion/pkg/api"
	"github.com/olad5/caution-companion/pkg/utils/logger"
)

func main() {
	configurations := config.GetConfig(".env")
	ctx := context.Background()

	l := logger.Get(configurations)

	postgresConnection := data.StartPostgres(configurations.DatabaseUrl, l)
	if err := postgres.Migrate(ctx, postgresConnection); err != nil {
		log.Fatal("Error Migrating postgres", err)
	}

	defer postgresConnection.Close()

	userRepo, err := postgres.NewPostgresUserRepo(ctx, postgresConnection)
	if err != nil {
		log.Fatal("Error Initializing User Repo", err)
	}

	reportsRepo, err := postgres.NewPostgresReportRepo(ctx, postgresConnection)
	if err != nil {
		log.Fatal("Error Initializing Reports Repo", err)
	}

	redisCache, err := redis.New(ctx, configurations)
	if err != nil {
		log.Fatal("Error Initializing redisCache", err)
	}

	fileStore, err := cloudinary.NewCloudinaryFileStore(ctx, configurations)
	if err != nil {
		log.Fatal("Error Initializing fileStore", err)
	}

	appRouter := api.NewHttpRouter(ctx, userRepo, reportsRepo, fileStore, redisCache, configurations, l)

	port := configurations.Port
	server := &http.Server{Addr: ":" + port, Handler: loggingMiddleware.RequestLogger(appRouter, configurations)}
	go func() {
		message := "Server is running on port " + port
		fmt.Println(message)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("HTTP server ListenAndServe: %v", err)
		}
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exiting gracefully")
}
