package proxy

import (
	"log"
	"net/http/httputil"
	"net/url"

	"gateway/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterUsersProxy(r *gin.Engine, usersServiceURL *url.URL) {
	usersProxy := httputil.NewSingleHostReverseProxy(usersServiceURL)
	r.Any("/users/*path", middleware.AuthRequired(), func(c *gin.Context) {
		if userID, exists := c.Get("user_id"); exists {
			c.Request.Header.Set("X-User-ID", userID.(string))
		}
		path := c.Param("path")
		if path == "" {
			path = "/"
		}
		c.Request.URL.Path = path
		log.Printf("Proxying request to users service: %s %s", c.Request.Method, path)
		usersProxy.ServeHTTP(c.Writer, c.Request)
	})
}
