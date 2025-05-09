package proxy

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func RegisterAuthProxy(r *gin.Engine, authServiceURL *url.URL) {
	authProxy := httputil.NewSingleHostReverseProxy(authServiceURL)
	r.Any("/auth/*path", func(c *gin.Context) {
		path := c.Param("path")
		if path == "" {
			path = "/"
		}
		c.Request.URL.Path = path
		authProxy.ServeHTTP(c.Writer, c.Request)
	})
}
