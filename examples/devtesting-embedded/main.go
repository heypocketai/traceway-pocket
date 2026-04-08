package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	tracewaybackend "github.com/tracewayapp/traceway/backend"
	traceway "go.tracewayapp.com"
)

//go:embed index.html
var indexHTML []byte

//go:embed cdn.html
var cdnHTML []byte

//go:embed static/*
var staticFS embed.FS

const (
	appPort         = 8080
	backendToken    = "backend-dev-token"
	frontendToken   = "frontend-dev-token"
	monitoringToken = "monitoring-dev-token"
)

func main() {
	go tracewaybackend.Run(
		tracewaybackend.WithSQLitePath("./storage/traceway.db"),
		tracewaybackend.WithPort(8082),
		tracewaybackend.WithDefaultUser("admin@localhost.com", "admin"),
		tracewaybackend.WithDefaultProject("Backend API", "go", backendToken),
		tracewaybackend.WithDefaultProject("jQuery Frontend", "jquery", frontendToken),
		tracewaybackend.WithDefaultProject("Traceway Monitoring", "go", monitoringToken),
		tracewaybackend.WithMonitoringURL(monitoringToken+"@http://localhost:8082/api/report"),
	)

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, traceway-trace-id")
		c.Header("Access-Control-Expose-Headers", "traceway-trace-id")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// router.Use(tracewaygin.New(
	// 	backendToken+"@http://localhost:8082/api/report",
	// 	tracewaygin.WithOnErrorRecording(tracewaygin.RecordingQuery|tracewaygin.RecordingBody|tracewaygin.RecordingHeader|tracewaygin.RecordingUrl),
	// ))

	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	router.GET("/cdn", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", cdnHTML)
	})

	staticSub, _ := fs.Sub(staticFS, "static")
	router.StaticFS("/static", http.FS(staticSub))

	router.GET("/api/test-error", func(c *gin.Context) {
		time.Sleep(time.Duration(20+rand.IntN(80)) * time.Millisecond)
		err := errors.New("simulated backend error for distributed trace testing")
		traceway.CaptureExceptionWithContext(c.Request.Context(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	})

	router.GET("/api/test-success", func(c *gin.Context) {
		time.Sleep(time.Duration(20+rand.IntN(80)) * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	fmt.Println()
	fmt.Println("=================================================")
	fmt.Printf("  Node build:  http://localhost:%d\n", appPort)
	fmt.Printf("  CDN (no build): http://localhost:%d/cdn\n", appPort)
	fmt.Println("  Dashboard:   http://localhost:8082")
	fmt.Println("  Login:       admin@localhost.com / admin")
	fmt.Println("=================================================")
	fmt.Println()

	router.Run(fmt.Sprintf(":%d", appPort))
}
