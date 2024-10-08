package api

import (
	"context"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/olad5/caution-companion/config"
	authMiddleware "github.com/olad5/caution-companion/internal/handlers/auth"
	fileHandlers "github.com/olad5/caution-companion/internal/handlers/files"
	reportsHandlers "github.com/olad5/caution-companion/internal/handlers/reports"
	userHandlers "github.com/olad5/caution-companion/internal/handlers/users"
	"github.com/olad5/caution-companion/internal/infra"
	"github.com/olad5/caution-companion/internal/services/auth"
	"github.com/olad5/caution-companion/internal/usecases/files"
	"github.com/olad5/caution-companion/internal/usecases/reports"
	"github.com/olad5/caution-companion/internal/usecases/users"
	response "github.com/olad5/caution-companion/pkg/utils"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

func NewHttpRouter(
	ctx context.Context,
	userRepo infra.UserRepository,
	reportsRepo infra.ReportRepository,
	fileStore infra.FileStore,
	cache infra.Cache,
	mailService infra.MailService,
	configurations *config.Configurations,
	l *zap.Logger,
) http.Handler {
	authService, err := auth.NewRedisAuthService(ctx, cache, configurations.JwtSecretKey)
	if err != nil {
		log.Fatal("Error Initializing Auth Service", err)
	}

	userService, err := users.NewUserService(userRepo, authService, mailService)
	if err != nil {
		log.Fatal("Error Initializing UserService")
	}

	userHandler, err := userHandlers.NewUserHandler(*userService, authService, l)
	if err != nil {
		log.Fatal("failed to create the User handler: ", err)
	}
	reportsService, err := reports.NewReportsService(reportsRepo)
	if err != nil {
		log.Fatal("Error Initializing UserService")
	}
	reportsHandler, err := reportsHandlers.NewReportsHandler(*reportsService, l)
	if err != nil {
		log.Fatal("failed to create the Report handler: ", err)
	}

	filesService, err := files.NewFileService(fileStore)
	if err != nil {
		log.Fatal("Error Initializing FilesService")
	}
	filesHandler, err := fileHandlers.NewFilesHandler(*filesService, l)
	if err != nil {
		log.Fatal("failed to create the User handler: ", err)
	}

	router := chi.NewRouter()

	// -------------------------------------------------------------------------
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// -------------------------------------------------------------------------
	// handler for documentation
	router.Handle("/doc.json", http.FileServer(http.Dir(filepath.Join("docs"))))
	router.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/doc.json"),
	))

	// -------------------------------------------------------------------------
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		isDbHealthy := true
		isCacheHealthy := true

		if err := userRepo.Ping(ctx); err != nil {
			l.Error("[HEALTH_CHECK]: ", zap.Error(err))
			isDbHealthy = false
		}
		if err := cache.Ping(ctx); err != nil {
			l.Error("[HEALTH_CHECK]: ", zap.Error(err))
			isCacheHealthy = false
		}

		if !(isDbHealthy && isCacheHealthy) {
			response.ErrorResponse(w, "server is down", http.StatusInternalServerError)
			return
		}
		response.SuccessResponse(w, "server is healthy", nil, l)
	})

	// -------------------------------------------------------------------------
	// TODO:TODO: add prefixes to these routes
	router.Group(func(r chi.Router) {
		r.Use(
			middleware.AllowContentType("application/json"),
			middleware.SetHeader("Content-Type", "application/json"),
		)
		r.Post("/users", userHandler.CreateUser)
		r.Post("/users/login", userHandler.Login)
		r.Post("/users/token/refresh", userHandler.RefreshAccessToken)
		r.Post("/users/forgot-password", userHandler.ForgotPassword)
		r.Post("/users/reset-password/verify-token", userHandler.VerifyResetPasswordToken)
		r.Post("/users/reset-password", userHandler.ResetPassword)
	})

	// -------------------------------------------------------------------------

	router.Group(func(r chi.Router) {
		r.Use(
			middleware.AllowContentType("application/json"),
			middleware.SetHeader("Content-Type", "application/json"),
		)
		r.Use(authMiddleware.EnsureAuthenticated(authService))

		r.Put("/users", userHandler.EditUser)
		r.Get("/users/me", userHandler.GetLoggedInUser)
		r.Put("/users/password", userHandler.ChangePassword)
	})

	router.Group(func(r chi.Router) {
		r.Use(
			middleware.AllowContentType("application/json"),
			middleware.SetHeader("Content-Type", "application/json"),
		)
		r.Use(authMiddleware.EnsureAuthenticated(authService))

		r.Post("/reports", reportsHandler.CreateReport)
		r.Get("/reports/{id}", reportsHandler.GetReportByReportId)
		r.Get("/reports/latest", reportsHandler.GetLatestReports)
	})

	router.Group(func(r chi.Router) {
		r.Use(middleware.AllowContentType("multipart/form-data"))
		r.Use(authMiddleware.EnsureAuthenticated(authService))

		r.Post("/files/upload", filesHandler.Upload)
	})

	return router
}
