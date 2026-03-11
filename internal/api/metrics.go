package api

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	GatewayRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total number of requests handled by the API Gateway",
		},
		[]string{"status"},
	)
)

func init() {
	prometheus.MustRegister(GatewayRequestsTotal)
}
