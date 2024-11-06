package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (a *applicationDependences) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//defer will be called when the stack unwinds
		defer func() {
			//recover from panic
			err := recover()
			if err != nil {
				w.Header().Set("Connection", "Close")
				a.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *applicationDependences) rateLimiting(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time //to remove unactive map entries
	}
	var mu sync.Mutex
	var clients = make(map[string]*client)
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock() //finish cleanup
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				a.serverErrorResponse(w, r, err)
				return
			}
			mu.Lock()
			_, found := clients[ip]
			if !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(a.config.limiter.rps), a.config.limiter.burst)}
			}
			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				a.rateLimitExceededResponse(w, r)
				return
			}
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}
