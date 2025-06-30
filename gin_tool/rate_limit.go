package gin_tool

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/ulule/limiter"
	memory_store "github.com/ulule/limiter/drivers/store/memory"
	redis_store "github.com/ulule/limiter/drivers/store/redis"
)

type RateLimitTool struct{}

func (t RateLimitTool) Middleware(
	store limiter.Store, rateLimit int64, period int,
	identifierBuilder func(c *gin.Context) string,
) gin.HandlerFunc {
	rate := limiter.Rate{
		Limit:  rateLimit,
		Period: time.Duration(period) * time.Second,
	}
	limiterInstance := limiter.New(store, rate)

	return func(c *gin.Context) {
		identifier := identifierBuilder(c)
		context, err := limiterInstance.Get(c, identifier)

		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				ErrorResponse(
					http.StatusInternalServerError,
					fmt.Sprintf("Rate limit error: %v", err),
					&struct{}{},
				))
			return
		}

		// set response headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rate.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

		// check if the rate limit is reached
		if context.Reached {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, CommonResponse[struct{}]{
				Code:    http.StatusTooManyRequests,
				Message: http.StatusText(http.StatusTooManyRequests),
			})
			return
		}

	}
}

func (t RateLimitTool) MiddlewareWithMemoryStore(
	rateLimit int64, period int,
	identifierBuilder func(c *gin.Context) string,
) gin.HandlerFunc {
	store := memory_store.NewStore()
	// store, err := memory_store.NewStoreWithOptions(limiter.StoreOptions{
	// 	Prefix: "rate_limit", // Memory store prefix
	// })
	return t.Middleware(store, rateLimit, period, identifierBuilder)
}

func (t RateLimitTool) MiddlewareWithRedisStore(
	redisClient *redis.Client, rateLimit int64, period int,
	identifierBuilder func(c *gin.Context) string,
) gin.HandlerFunc {
	store, err := redis_store.NewStore(redisClient)
	// store, err := redis_store.NewStoreWithOptions(redisClient, limiter.StoreOptions{
	// 	Prefix: "rate_limit", // Redis 键前缀
	// })
	if err != nil {
		panic(fmt.Sprintf("Failed to create Redis store: %v", err))
	}
	return t.Middleware(store, rateLimit, period, identifierBuilder)
}

// DefaultIdentifierBuilder is a default implementation of the identifier builder function.
// It generates an identifier based on the client's IP address, request method, and full path.
func DefaultIdentifierBuilder(c *gin.Context) string {
	return fmt.Sprintf(
		"%s:%s:%s:%s",
		"rate_limit",
		c.ClientIP(),
		c.Request.Method,
		c.FullPath(),
	)
}
