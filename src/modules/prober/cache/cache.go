package cache

import (
	"context"

	"github.com/Major818/nightingale/v4/src/modules/prober/config"
)

var CollectRule *CollectRuleCache // collectrule.go
var MetricHistory *history        // history.go

func Init(ctx context.Context) error {
	CollectRule = NewCollectRuleCache(&config.Config.CollectRule)
	CollectRule.start(ctx)
	MetricHistory = NewHistory()
	config.InitPluginsConfig(config.Config)
	return nil
}
