package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"gateway/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func RegisterTournamentProxy(r *gin.Engine, tournamentServiceURL *url.URL, logger *logrus.Logger) {
	proxy := httputil.NewSingleHostReverseProxy(tournamentServiceURL)

	proxy.Director = func(req *http.Request) {
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/tournaments")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}

		req.URL.Scheme = tournamentServiceURL.Scheme
		req.URL.Host = tournamentServiceURL.Host
		req.Host = tournamentServiceURL.Host

		if clientIP := req.Header.Get("X-Forwarded-For"); clientIP == "" {
			req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		http.Error(w, "tournament service unavailable: "+err.Error(), http.StatusBadGateway)
	}

	r.Any("/tournaments/*path", middleware.AuthRequired(logger), func(c *gin.Context) {
		// Headers
		if userID, ok := c.Get("user_id"); ok {
			c.Request.Header.Set("X-User-ID", userID.(string))
		}
		if role, ok := c.Get("user_role"); ok {
			c.Request.Header.Set("X-User-Role", role.(string))
		}
		if reqID, ok := c.Get("request_id"); ok {
			c.Request.Header.Set("X-Request-ID", reqID.(string))
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	})
}
