package mysql

import (
	"os"
	"testing"

	"github.com/Major818/n9e/v4/src/modules/server/plugins"
)

func TestCollect(t *testing.T) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Error("unable to get DATA_SOURCE_NAME from environment")
	}

	plugins.PluginTest(t, &MysqlRule{
		Servers:           []string{dsn},
		GatherSlaveStatus: true,
	})
}
