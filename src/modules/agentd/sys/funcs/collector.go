package funcs

import (
	"github.com/Major818/n9e/v4/src/common/dataobj"
	"github.com/Major818/n9e/v4/src/modules/agentd/core"
)

func CollectorMetrics() []*dataobj.MetricValue {
	return []*dataobj.MetricValue{
		core.GaugeValue("proc.agent.alive", 1),
	}
}
