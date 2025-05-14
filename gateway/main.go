package main

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"gateway/proxy"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	log.Println("GATEWAY JWT_SECRET:", os.Getenv("JWT_SECRET"))

	r := gin.Default()

	// Middleware: CORS
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-User-Role")

		// Обработка preflight запроса
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Middleware: simple request log
	r.Use(func(c *gin.Context) {
		log.Printf("[%s] %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	logger := logrus.New()

	// Parse service URLs
	authServiceURL := mustParse("AUTH_SERVICE_URL")
	usersServiceURL := mustParse("USERS_SERVICE_URL")
	gameServiceURL := mustParse("GAME_SERVICE_URL")
	tournamentServiceURL := mustParse("TOURNAMENT_SERVICE_URL")

	// Register proxies
	proxy.RegisterAuthProxy(r, authServiceURL, logger)
	proxy.RegisterUsersProxy(r, usersServiceURL, logger)
	proxy.RegisterGameProxy(r, gameServiceURL, logger)
	proxy.RegisterTournamentProxy(r, tournamentServiceURL, logger)

	// Start server
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Gateway server running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func mustParse(env string) *url.URL {
	raw := os.Getenv(env)
	if raw == "" {
		log.Fatalf("missing %s", env)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		log.Fatalf("invalid %s: %v", env, err)
	}
	return parsed
}
