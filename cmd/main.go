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

	"github.com/olad5/go-hackathon-starter-template/config/data"
	"github.com/olad5/go-hackathon-starter-template/internal/infra/postgres"
	"github.com/olad5/go-hackathon-starter-template/internal/infra/redis"

	"github.com/olad5/go-hackathon-starter-template/config"
	loggingMiddleware "github.com/olad5/go-hackathon-starter-template/internal/handlers/logging"
	"github.com/olad5/go-hackathon-starter-template/pkg/api"
	"github.com/olad5/go-hackathon-starter-template/pkg/utils/logger"
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

	redisCache, err := redis.New(ctx, configurations)
	if err != nil {
		log.Fatal("Error Initializing redisCache", err)
	}

	appRouter := api.NewHttpRouter(ctx, userRepo, redisCache, configurations, l)

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
