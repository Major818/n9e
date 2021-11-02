package dns_query

import (
	"testing"

	"github.com/Major818/nightingale/v4/src/modules/server/plugins"
)

func TestCollect(t *testing.T) {
	plugins.PluginTest(t, &Rule{
		Servers: []string{"223.5.5.5"},
		Domains: []string{"www.baidu.com"},
	})
}
