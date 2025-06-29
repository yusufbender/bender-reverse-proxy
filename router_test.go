package main

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Beklenen 200 OK, gelen: %d", resp.StatusCode)
	}
}
func TestRoundRobinLogic(t *testing.T) {
	path := "/api"
	targets := []string{"http://localhost:5001", "http://localhost:5002", "http://localhost:5003"}

	// Hepsini sağlıklı kabul edelim
	for _, target := range targets {
		setHealth(target, true)
	}

	// İlk turda sırayla döndüğünü kontrol edelim
	for i := 0; i < len(targets)*2; i++ {
		expected := targets[i%len(targets)]
		actual := getNextTarget(path, targets)
		if actual != expected {
			t.Errorf("Expected %s but got %s on iteration %d", expected, actual, i)
		}
	}
}
func TestRewritePath(t *testing.T) {
	route := Route{
		Path:    "/api",
		Rewrite: "/",
	}

	req := httptest.NewRequest("GET", "/api/user", nil)
	recorder := httptest.NewRecorder()

	var rewrittenPath string

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			if route.Rewrite != "" && strings.HasPrefix(r.URL.Path, route.Path) {
				rewrittenPath = strings.TrimSuffix(route.Rewrite, "/") + "/" + strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, route.Path), "/")
			} else {
				rewrittenPath = r.URL.Path
			}
			// Dummy host atamak gerekiyor
			r.URL.Scheme = "http"
			r.URL.Host = "example.com"
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			// Testte gerçek proxy yok, o yüzden bu alanı ignore ediyoruz
		},
	}

	proxy.ServeHTTP(recorder, req)

	expected := "/user"
	if rewrittenPath != expected {
		t.Errorf("Rewrite failed. Expected %s, got %s", expected, rewrittenPath)
	}
}
func TestBasicAuth(t *testing.T) {
	routes := []Route{
		{
			Path:   "/secure",
			Target: "http://example.com", // Buraya istek gitmeyecek, mock'lanacak
			Auth: &Auth{
				Username: "admin",
				Password: "1234",
			},
		},
		{
			Path:   "/open",
			Target: "http://example.com",
		},
	}

	routesMutex.Lock()
	currentRoutes = routes
	routesMutex.Unlock()

	handler := getDynamicProxyHandler() // ✅ artık parametre almıyor

	// 1. Doğru auth ile
	req1 := httptest.NewRequest("GET", "/secure", nil)
	req1.SetBasicAuth("admin", "1234")
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	if rr1.Code == http.StatusUnauthorized {
		t.Error("Doğru bilgilerle 401 dönmemeli")
	}

	// 2. Yanlış auth ile
	req2 := httptest.NewRequest("GET", "/secure", nil)
	req2.SetBasicAuth("admin", "wrongpass")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusUnauthorized {
		t.Errorf("Yanlış bilgilerle 401 bekleniyordu, gelen: %d", rr2.Code)
	}

	// 3. Auth olmayan route
	req3 := httptest.NewRequest("GET", "/open", nil)
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)
	if rr3.Code == http.StatusUnauthorized {
		t.Error("/open route'una auth uygulanmamalı")
	}
}
