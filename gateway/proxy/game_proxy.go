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

func RegisterGameProxy(r *gin.Engine, gameServiceURL *url.URL, logger *logrus.Logger) {
	proxy := httputil.NewSingleHostReverseProxy(gameServiceURL)

	// Кастомный Director
	proxy.Director = func(req *http.Request) {
		// Удаляем "/game" префикс из пути
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/game")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}

		req.URL.Scheme = gameServiceURL.Scheme
		req.URL.Host = gameServiceURL.Host
		req.Host = gameServiceURL.Host

		// Прокидываем IP
		if clientIP := req.Header.Get("X-Forwarded-For"); clientIP == "" {
			req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		}
	}

	// Обработка ошибок прокси
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		http.Error(w, "game service unavailable: "+err.Error(), http.StatusBadGateway)
	}

	r.Any("/game/*path", middleware.AuthRequired(logger), func(c *gin.Context) {
		// Прокидываем user_id
		if userID, exists := c.Get("user_id"); exists {
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
