package main

import (
	"log"
	"net/http"
	"time"
)

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}

		log.Printf("[%s] %s %s from %s", time.Now().Format(time.RFC1123), r.Method, r.URL.Path, ip)
		next.ServeHTTP(w, r)
	})
}
func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}

		limiter := getVisitor(ip)
		if !limiter.Allow() {
			http.Error(w, "429 - Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	const configPath = "routes.yaml"

	// Başlangıçta config dosyasını yükle
	routes, err := loadRoutes(configPath)
	if err != nil {
		log.Fatalf("Route dosyası yüklenemedi: %v", err)
	}
	currentRoutes = routes

	// Config dosyasını izlemeye başla
	go watchConfigFile(configPath, 2*time.Second)

	// Sağlık kontrollerini sürekli yap
	go func() {
		for {
			routesMutex.RLock()
			copied := make([]Route, len(currentRoutes))
			copy(copied, currentRoutes)
			routesMutex.RUnlock()

			healthCheckLoop(copied, 10*time.Second)
		}
	}()

	// Ziyaretçi temizliği
	go cleanupVisitors()

	// Middleware’leri sırayla bağla: logging > rateLimit > proxyHandler
	handler := loggingMiddleware(rateLimitMiddleware(http.HandlerFunc(getDynamicProxyHandler())))

	http.Handle("/", handler)

	log.Println("🚀 Dinamik Reverse Proxy başladı: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
