package main

import (
	"log"
	"net/http/httputil"
	"net/url"
	"os"

	"StartupPCConfigurator/pkg/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret_key"
	}

	router := gin.Default()
	router.Use(cors.Default())

	// 🔓 Публичный /auth/*
	router.Any("/auth/*proxyPath", reverseProxy("http://localhost:8001"))

	// 🔓 Публичные ручки config-сервиса
	router.GET("/config/components", reverseProxyPath("http://localhost:8002", "/components"))
	router.GET("/config/compatible", reverseProxyPath("http://localhost:8002", "/compatible"))

	// 🔐 Защищённые ручки config-сервиса через /config-secure/*
	configProtected := router.Group("/config-secure")
	configProtected.Use(middleware.AuthMiddleware(jwtSecret))
	{
		configProtected.Any("/*proxyPath", reverseProxy("http://localhost:8002"))
	}

	// 🔐 Защищённые ручки aggregator-сервиса через /offers/*
	offersGroup := router.Group("/offers")
	offersGroup.Use(middleware.AuthMiddleware(jwtSecret))
	{
		offersGroup.Any("/*proxyPath", reverseProxy("http://localhost:8003"))
	}

	log.Println("🚀 Gateway запущен на :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("❌ Не удалось запустить gateway: %v", err)
	}
}

// reverseProxy — для маршрутов с *proxyPath
func reverseProxy(target string) gin.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Невалидный адрес сервиса: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)

	return func(c *gin.Context) {
		c.Request.URL.Path = c.Param("proxyPath")
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// reverseProxyPath — для фиксированных путей без wildcard
func reverseProxyPath(target, path string) gin.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Невалидный адрес сервиса: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)

	return func(c *gin.Context) {
		c.Request.URL.Path = path
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
