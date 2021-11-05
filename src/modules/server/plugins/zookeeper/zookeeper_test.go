package zookeeper

import (
	"testing"
	"time"

	"github.com/Major818/n9e/v4/src/modules/server/plugins"
)

func TestCollect(t *testing.T) {
	input := plugins.PluginTest(t, &Rule{
		Servers: []string{"localhost:2181"},
	})

	time.Sleep(time.Second)
	plugins.PluginInputTest(t, input)
}
