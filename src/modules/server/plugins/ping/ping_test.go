package ping

import (
	"testing"

	"github.com/Major818/n9e/v4/src/modules/server/plugins"
)

func TestCollect(t *testing.T) {
	plugins.PluginTest(t, &Rule{
		Urls: []string{"github.com", "n9e.didiyun.com"},
	})
}
