package nginx

import (
	"testing"
	"time"

	"github.com/Major818/n9e/v4/src/modules/server/plugins"
)

func TestCollect(t *testing.T) {
	input := plugins.PluginTest(t, &Rule{
		Urls: []string{"http://localhost/nginx-status"},
	})

	time.Sleep(time.Second)
	plugins.PluginInputTest(t, input)
}
