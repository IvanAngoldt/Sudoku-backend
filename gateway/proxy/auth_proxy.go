package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func RegisterAuthProxy(r *gin.Engine, authServiceURL *url.URL, logger *logrus.Logger) {
	proxy := httputil.NewSingleHostReverseProxy(authServiceURL)

	// Кастомный director
	proxy.Director = func(req *http.Request) {
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/auth")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}

		req.URL.Scheme = authServiceURL.Scheme
		req.URL.Host = authServiceURL.Host
		req.Host = authServiceURL.Host

		if clientIP := req.Header.Get("X-Forwarded-For"); clientIP == "" {
			req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		}
	}

	// Обработка ошибок прокси
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		logger.WithFields(logrus.Fields{
			"target": authServiceURL.String(),
			"error":  err.Error(),
		}).Error("auth service unavailable")
		http.Error(rw, "auth service unavailable: "+err.Error(), http.StatusBadGateway)
	}

	// Прокси без AuthRequired
	r.Any("/auth/*path", func(c *gin.Context) {
		if reqID, ok := c.Get("request_id"); ok {
			c.Request.Header.Set("X-Request-ID", reqID.(string))
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	})
}
