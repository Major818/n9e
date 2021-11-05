package http

import (
	"github.com/Major818/n9e/v4/src/modules/server/config"

	"github.com/gin-gonic/gin"
)

func ldapUsed(c *gin.Context) {
	renderData(c, config.Config.Rdb.LDAP.DefaultUse, nil)
}
