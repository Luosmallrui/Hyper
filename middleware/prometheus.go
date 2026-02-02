package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// 定义一个 Counter 类型的指标，用于统计 QPS (请求次数)
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "my_app_http_requests_total", // 指标名称，在阿里云里搜这个
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"}, // 维度标签
	)

	// 定义一个 Histogram 类型的指标，用于统计响应耗时
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "my_app_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5}, // 统计耗时分布的区间
		},
		[]string{"method", "path"},
	)
)

func init() {
	// 注册指标到默认注册表
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath() // 获取路由路径
		if path == "" {
			path = "unknown"
		}

		c.Next() // 执行实际业务逻辑

		// 记录指标数据
		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
