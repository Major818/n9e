package procs

import (
	"time"

	"github.com/Major818/nightingale/v4/src/modules/agentd/stra"
)

func Detect() {
	detect()
	go loopDetect()
}

func loopDetect() {
	for {
		time.Sleep(time.Second * 10)
		detect()
	}
}

func detect() {
	ps := stra.GetProcCollects()
	DelNoProcCollect(ps)
	AddNewProcCollect(ps)
}
