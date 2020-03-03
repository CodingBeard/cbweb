package cbweb

import (
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/config"
	"time"
)

type LimiterConfig struct {
	Max int64
	Ttl time.Duration
	Methods []string
	ContentType string
	Message string
	ForwardedForIp bool
}

func NewLimiter(config LimiterConfig) *config.Limiter {
	if config.Max == 0 {
		config.Max = 1
	}
	if config.Ttl == 0 {
		config.Ttl = time.Second
	}
	if config.ContentType == "" {
		config.ContentType = "text/plain"
	}

	limiter := tollbooth.NewLimiterExpiringBuckets(config.Max, config.Ttl, time.Hour, time.Second)
	if len(config.Methods) != 0 {
		limiter.Methods = config.Methods
	}
	limiter.MessageContentType = config.ContentType
	limiter.Message = config.Message
	if config.ForwardedForIp {
		limiter.IPLookups = []string{"X-Forwarded-For", "RemoteAddr"}
	}

	return limiter
}