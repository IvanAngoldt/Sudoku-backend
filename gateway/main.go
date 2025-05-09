package main

import (
	"log"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"gateway/proxy"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("GATEWAY JWT_SECRET:", os.Getenv("JWT_SECRET"))

	r := gin.Default()

	// Настройка CORS
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Глобальный middleware для логирования
	r.Use(func(c *gin.Context) {
		log.Printf("%s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	authServiceURL, _ := url.Parse(os.Getenv("AUTH_SERVICE_URL"))
	usersServiceURL, _ := url.Parse(os.Getenv("USERS_SERVICE_URL"))
	gameServiceURL, _ := url.Parse(os.Getenv("GAME_SERVICE_URL"))
	tournamentServiceURL, _ := url.Parse(os.Getenv("TOURNAMENT_SERVICE_URL"))

	proxy.RegisterAuthProxy(r, authServiceURL)
	proxy.RegisterUsersProxy(r, usersServiceURL)
	proxy.RegisterGameProxy(r, gameServiceURL)
	proxy.RegisterTournamentProxy(r, tournamentServiceURL)

	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Gateway server is running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
