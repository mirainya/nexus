package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	global  *rate.Limiter
	ips     sync.Map
	ipRate  rate.Limit
	ipBurst int
}

func NewRateLimiter(globalRate rate.Limit, globalBurst int, ipRate rate.Limit, ipBurst int) *RateLimiter {
	rl := &RateLimiter{
		global:  rate.NewLimiter(globalRate, globalBurst),
		ipRate:  ipRate,
		ipBurst: ipBurst,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getIPLimiter(ip string) *rate.Limiter {
	v, ok := rl.ips.Load(ip)
	if ok {
		entry := v.(*ipLimiter)
		entry.lastSeen = time.Now()
		return entry.limiter
	}
	limiter := rate.NewLimiter(rl.ipRate, rl.ipBurst)
	rl.ips.Store(ip, &ipLimiter{limiter: limiter, lastSeen: time.Now()})
	return limiter
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.ips.Range(func(key, value any) bool {
			entry := value.(*ipLimiter)
			if time.Since(entry.lastSeen) > 3*time.Minute {
				rl.ips.Delete(key)
			}
			return true
		})
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.global.Allow() {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "server rate limit exceeded"})
			return
		}
		if !rl.getIPLimiter(c.ClientIP()).Allow() {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "ip rate limit exceeded"})
			return
		}
		c.Next()
	}
}
