package proxy

import (
	"log"
	"net/http/httputil"
	"net/url"

	"gateway/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterTournamentProxy(r *gin.Engine, tournamentServiceURL *url.URL) {
	tournamentProxy := httputil.NewSingleHostReverseProxy(tournamentServiceURL)
	r.Any("/tournaments/*path", middleware.AuthRequired(), func(c *gin.Context) {
		if userID, exists := c.Get("user_id"); exists {
			c.Request.Header.Set("X-User-ID", userID.(string))
		}
		path := c.Param("path")
		if path == "" {
			path = "/"
		}
		c.Request.URL.Path = path
		log.Printf("Proxying request to tournament service: %s %s", c.Request.Method, path)
		tournamentProxy.ServeHTTP(c.Writer, c.Request)
	})
}
