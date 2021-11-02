package demo

import (
	"testing"

	"github.com/Major818/nightingale/v4/src/modules/server/plugins"
)

func TestCollect(t *testing.T) {
	plugins.PluginTest(t, &DemoRule{
		Period: 3600,
		Count:  10,
	})
}
