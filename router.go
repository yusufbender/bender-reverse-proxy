package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	rrCounters     = make(map[string]int)
	rrMutex        sync.Mutex
	healthyTargets = make(map[string]bool) // her hedefin sağlık durumu

	currentRoutes []Route
	routesMutex   sync.RWMutex
)

type Auth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Route struct {
	Path    string   `yaml:"path"`
	Target  string   `yaml:"target,omitempty"`
	Targets []string `yaml:"targets,omitempty"`
	Auth    *Auth    `yaml:"auth,omitempty"`
	Rewrite string   `yaml:"rewrite,omitempty"` // 👈 yeni alan
}

type Config struct {
	Routes []Route `yaml:"routes"`
}

func loadRoutes(configPath string) ([]Route, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config.Routes, nil
}

func getDynamicProxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		routesMutex.RLock()
		routes := make([]Route, len(currentRoutes))
		copy(routes, currentRoutes)
		routesMutex.RUnlock()

		for _, route := range routes {
			if strings.HasPrefix(r.URL.Path, route.Path) {
				if route.Auth != nil {
					user, pass, ok := r.BasicAuth()
					if !ok || user != route.Auth.Username || pass != route.Auth.Password {
						w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
						http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
						return
					}
				}

				if len(route.Targets) > 0 {
					target := getNextTarget(route.Path, route.Targets)
					proxyTo(target, route, w, r)
					return
				}

				if route.Target != "" {
					proxyTo(route.Target, route, w, r)
					return
				}
			}
		}

		http.Error(w, "404 - Eşleşen route yok", http.StatusNotFound)
	}
}

func getNextTarget(path string, targets []string) string {
	rrMutex.Lock()
	defer rrMutex.Unlock()

	// SAĞLIKLI hedefler süzülüyor
	var available []string
	for _, t := range targets {
		if healthyTargets[t] {
			available = append(available, t)
		}
	}

	if len(available) == 0 {
		// Hepsi çökmüşse fallback: tüm hedeflere izin ver
		available = targets
	}

	i := rrCounters[path]
	target := available[i%len(available)]
	rrCounters[path] = i + 1
	return target
}

func proxyTo(targetURL string, route Route, w http.ResponseWriter, r *http.Request) {
	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Geçersiz hedef URL", http.StatusInternalServerError)
		return
	}

	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		// Rewrite varsa uygula
		if route.Rewrite != "" && strings.HasPrefix(req.URL.Path, route.Path) {
			rewrittenPath := strings.Replace(req.URL.Path, route.Path, route.Rewrite, 1)
			req.URL.Path = rewrittenPath
		} else {
			req.URL.Path = r.URL.Path
		}

		req.URL.RawQuery = r.URL.RawQuery
	}

	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(w, r)
}
func healthCheckLoop(routes []Route, interval time.Duration) {
	for {
		for _, route := range routes {
			// Tüm target’ları sırayla kontrol et
			var targets []string
			if len(route.Targets) > 0 {
				targets = route.Targets
			} else if route.Target != "" {
				targets = []string{route.Target}
			}

			for _, target := range targets {
				go checkTargetHealth(target)
			}
		}
		time.Sleep(interval)
	}
}

func checkTargetHealth(target string) {
	resp, err := http.Get(target + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		setHealth(target, false)
		return
	}
	setHealth(target, true)
}

func setHealth(target string, status bool) {
	rrMutex.Lock()
	defer rrMutex.Unlock()
	healthyTargets[target] = status
}
func watchConfigFile(path string, interval time.Duration) {
	var lastModTime time.Time

	for {
		fi, err := os.Stat(path)
		if err != nil {
			log.Printf("Config dosyası okunamadı: %v\n", err)
			time.Sleep(interval)
			continue
		}

		modTime := fi.ModTime()
		if modTime.After(lastModTime) {
			log.Println("🌀 Config dosyası değişti, yeniden yükleniyor...")

			routes, err := loadRoutes(path)
			if err != nil {
				log.Printf("Config yükleme hatası: %v\n", err)
			} else {
				routesMutex.Lock()
				currentRoutes = routes
				routesMutex.Unlock()
				log.Println("✅ Config yeniden yüklendi.")
			}

			lastModTime = modTime
		}

		time.Sleep(interval)
	}
}
