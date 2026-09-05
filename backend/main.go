package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/korvlad21/firstGoWeb/handler"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.Default()
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatal(err)
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
	log.Fatal(server.ListenAndServe())
}
