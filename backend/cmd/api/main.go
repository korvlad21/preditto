package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"preditto/internal/config"
	"preditto/internal/database"
	"preditto/internal/handler"
)

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

	router := gin.Default()
	if err := router.SetTrustedProxies(nil); err != nil {
		return err
	}

	helloHandler := handler.NewHelloHandler()

	router.GET("/", gin.WrapF(helloHandler.Index))
	router.GET("/hello", gin.WrapF(helloHandler.Hello))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("HTTP server listening on %s", server.Addr)
	return server.ListenAndServe()
}
