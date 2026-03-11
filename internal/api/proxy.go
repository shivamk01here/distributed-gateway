// File: internal/api/proxy.go
package api

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type LoadBalancer struct {
	backends []*url.URL
	current  uint64
}

func NewLoadBalancer(targetHosts []string) *LoadBalancer {
	var urls []*url.URL
	for _, host := range targetHosts {
		u, err := url.Parse(host)
		if err != nil {
			log.Fatalf("Invalid backend URL %s: %v", host, err)
		}
		urls = append(urls, u)
	}

	return &LoadBalancer{
		backends: urls,
		current:  0,
	}
}

func (lb *LoadBalancer) getNextBackend() *url.URL {
	next := atomic.AddUint64(&lb.current, 1)

	index := (next - 1) % uint64(len(lb.backends))

	return lb.backends[index]
}

func (lb *LoadBalancer) ReverseProxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := lb.getNextBackend()

		proxy := httputil.NewSingleHostReverseProxy(target)

		r.URL.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = target.Host

		log.Printf("[PROXY] Load Balancing request to -> %s", target.Host)
		proxy.ServeHTTP(w, r)
	}
}
