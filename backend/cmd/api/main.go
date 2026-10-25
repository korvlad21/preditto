package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"preditto/internal/auth"
	"preditto/internal/config"
	"preditto/internal/database"
	"preditto/internal/handler"
	"preditto/internal/middleware"
	"preditto/internal/repository"
	apirouter "preditto/internal/router"
	"preditto/internal/service"
)

// @title Preditto API
// @version 1.0
// @description REST API for authentication, current users, teams, countries and tournaments. /api responses use success/data or success/error envelopes. Empty success omits data; empty lists are []; nullable fields stay null. Unsupported API methods/paths return 404 NOT_FOUND. Demo and health routes use their own response shapes.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter Bearer followed by a space and your access_token from register/login. Refresh tokens are accepted only in refresh/logout JSON bodies.
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	db, err := database.NewPostgres(ctx, cfg.Postgres)
	cancel()
	if err != nil {
		return err
	}
	defer db.Close()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	router, err := newRouter(db, cfg)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("HTTP server listening on %s", server.Addr)
	return server.ListenAndServe()
}

// newRouter keeps route composition shared by the server and Swagger coverage tests.
func newRouter(db *sql.DB, cfg config.Config) (*gin.Engine, error) {
	router := apirouter.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	helloHandler := handler.NewHelloHandler()
	tokens, err := auth.NewTokenManager(cfg.Auth)
	if err != nil {
		return nil, err
	}
	users := repository.NewUserRepository(db)
	authService, err := service.NewAuthService(repository.NewAuthRepository(db), users, tokens)
	if err != nil {
		return nil, err
	}
	apirouter.RegisterAuthRoutes(router.Group("/api"), handler.NewAuthHandler(authService))
	userService := service.NewUserService(users)
	authMiddleware := []gin.HandlerFunc{middleware.RequireAuth(tokens), middleware.LoadUser(userService)}
	protected := router.Group("/api", authMiddleware...)
	apirouter.RegisterUserRoutes(protected, handler.NewUserHandler())
	tournamentService := service.NewTournamentService(repository.NewTournamentRepository(db), users)
	apirouter.RegisterTournamentRoutes(protected, handler.NewTournamentHandler(tournamentService))
	teamService := service.NewTeamService(repository.NewTeamRepository(db))
	teamHandler := handler.NewTeamHandler(teamService)
	countryService := service.NewCountryService(repository.NewCountryRepository(db))
	apirouter.RegisterTeamRoutes(router.Group("/api"), teamHandler, handler.NewCountryHandler(countryService))
	apirouter.RegisterProtectedTeamRoutes(protected, teamHandler)

	router.GET("/", gin.WrapF(helloHandler.Index))
	router.GET("/hello", gin.WrapF(helloHandler.Hello))
	router.GET("/health", health)
	apirouter.RegisterSwaggerRoutes(router, cfg.Environment)
	return router, nil
}

// health reports process responsiveness, not database readiness.
// @Summary Process health
// @Tags System
// @Produce json
// @Success 200 {object} apidocs.Health
// @Router /health [get]
func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
