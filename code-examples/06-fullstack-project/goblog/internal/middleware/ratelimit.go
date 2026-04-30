package middleware

import (
	"sync"

	"guide-go/goblog/internal/config"
	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipLimiter 按 IP 地址管理令牌桶限流器
type ipLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

// newIPLimiter 创建 IP 限流器管理器
func newIPLimiter(r float64, burst int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(r),
		burst:    burst,
	}
}

// getLimiter 获取指定 IP 的限流器，不存在则创建
func (l *ipLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.RLock()
	limiter, exists := l.limiters[ip]
	l.mu.RUnlock()

	if exists {
		return limiter
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 双重检查
	if limiter, exists = l.limiters[ip]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(l.rate, l.burst)
	l.limiters[ip] = limiter
	return limiter
}

// RateLimiter 返回令牌桶限流中间件
// 按客户端 IP 进行限流，超过限制返回 429 Too Many Requests
func RateLimiter(cfg *config.RateLimitConfig) gin.HandlerFunc {
	limiter := newIPLimiter(cfg.RequestsPerSecond, cfg.Burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		l := limiter.getLimiter(ip)

		if !l.Allow() {
			response.Error(c, errcode.ErrTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}
