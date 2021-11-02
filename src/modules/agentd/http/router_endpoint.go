package http

import (
	"github.com/Major818/nightingale/v4/src/modules/agentd/config"

	"github.com/gin-gonic/gin"
)

func endpoint(c *gin.Context) {
	c.String(200, config.Endpoint)
}
