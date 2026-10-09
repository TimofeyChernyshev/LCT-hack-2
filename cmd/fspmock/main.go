package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"TimofeyChernyshev/LCT-hack-2/pkg/fsp"
)

func main() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8088"
	}

	env := os.Getenv("APP_ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	registry := fsp.NewMockRegistry()

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Health check
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "fsp-registry-mock",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	api := router.Group("/api/v1/fsp/registry")
	{
		// Search members
		api.GET("/members", func(c *gin.Context) {
			query := c.Query("q")
			rank := fsp.SportsRank(c.Query("rank"))
			region := c.Query("region")
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
			offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

			members, total, err := registry.SearchMembers(c.Request.Context(), query, rank, region, limit, offset)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"total":   total,
				"members": members,
			})
		})

		// Get member by FSP ID
		api.GET("/members/:fsp_id", func(c *gin.Context) {
			fspID := c.Param("fsp_id")
			member, err := registry.GetMember(c.Request.Context(), fspID)
			if err != nil {
				if errors.Is(err, fsp.ErrMemberNotFound) {
					c.JSON(http.StatusNotFound, gin.H{"error": "member not found in fsp registry"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, member)
		})

		// Verify member credentials
		api.POST("/verify", func(c *gin.Context) {
			var req fsp.VerificationRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid verification request body"})
				return
			}

			res, err := registry.VerifyMember(c.Request.Context(), req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		// Dynamically register/add member (for testing)
		api.POST("/members", func(c *gin.Context) {
			var member fsp.Member
			if err := c.ShouldBindJSON(&member); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member json"})
				return
			}

			if err := registry.RegisterMember(c.Request.Context(), member); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, member)
		})
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("[fsp-registry-mock] starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[fsp-registry-mock] shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced to shutdown: %v", err)
	}
	log.Println("[fsp-registry-mock] stopped cleanly")
}
