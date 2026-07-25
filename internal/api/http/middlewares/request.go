package middlewares

import "github.com/gin-gonic/gin"

// RequestLogger — middleware-заглушка под логирование запросов.
func RequestLogger(c *gin.Context) {
	c.Next()
}
