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

func ProxyWithUserHeaders(proxy *httputil.ReverseProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
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
	}
}

func RegisterUsersProxy(r *gin.Engine, usersServiceURL *url.URL, logger *logrus.Logger) {
	proxy := httputil.NewSingleHostReverseProxy(usersServiceURL)

	proxy.Director = func(req *http.Request) {
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/users")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
		req.URL.Scheme = usersServiceURL.Scheme
		req.URL.Host = usersServiceURL.Host
		req.Host = usersServiceURL.Host

		if clientIP := req.Header.Get("X-Forwarded-For"); clientIP == "" {
			req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		http.Error(w, "users service unavailable: "+err.Error(), http.StatusBadGateway)
	}

	// -------- 📣 Публичные маршруты --------
	public := r.Group("/users")
	public.Any("/check-username", gin.WrapH(proxy))
	public.Any("/check-email", gin.WrapH(proxy))
	public.POST("/", gin.WrapH(proxy))
	public.POST("/auth", gin.WrapH(proxy))

	// -------- 🔒 Защищённые маршруты с проксированием X-User-ID и др. --------
	protected := r.Group("/users", middleware.AuthRequired(logger))

	protected.GET("/", ProxyWithUserHeaders(proxy))
	protected.GET("/:id", ProxyWithUserHeaders(proxy))
	protected.PATCH("/:id", ProxyWithUserHeaders(proxy))
	protected.DELETE("/:id", ProxyWithUserHeaders(proxy))

	protected.GET("/me", ProxyWithUserHeaders(proxy))
	protected.GET("/me/info", ProxyWithUserHeaders(proxy))
	protected.POST("/me/info", ProxyWithUserHeaders(proxy))
	protected.PATCH("/me/info", ProxyWithUserHeaders(proxy))

	protected.POST("/me/avatar", ProxyWithUserHeaders(proxy))
	protected.GET("/:id/avatar", ProxyWithUserHeaders(proxy))

	protected.GET("/:id/statistics", ProxyWithUserHeaders(proxy))
	protected.PATCH("/:id/statistics", ProxyWithUserHeaders(proxy))
}
