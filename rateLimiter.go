package main

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*client)
	mu       sync.Mutex
	rps      = 2               // saniyede 2 istek
	burst    = 5               // ani patlamaya 5 tokenlık esneklik
	ttl      = time.Minute * 3 // 3 dakika aktif olmayanlar silinir
)

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	c, exists := visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(rps), burst)
		visitors[ip] = &client{limiter, time.Now()}
		return limiter
	}

	c.lastSeen = time.Now()
	return c.limiter
}

// Eski client’ları temizle (çöp toplama gibi)
func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, c := range visitors {
			if time.Since(c.lastSeen) > ttl {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}
