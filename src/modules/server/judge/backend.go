// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package judge

import (
	"strings"
	"time"

	"github.com/Major818/n9e/v4/src/common/dataobj"
	"github.com/Major818/n9e/v4/src/common/pools"
	"github.com/Major818/n9e/v4/src/common/stats"
	"github.com/Major818/n9e/v4/src/common/str"
	"github.com/Major818/n9e/v4/src/models"
	"github.com/Major818/n9e/v4/src/modules/server/cache"

	"github.com/toolkits/pkg/concurrent/semaphore"
	"github.com/toolkits/pkg/container/list"
	"github.com/toolkits/pkg/logger"
)

type JudgeSection struct {
	Batch       int    `yaml:"batch"`
	ConnTimeout int    `yaml:"connTimeout"`
	CallTimeout int    `yaml:"callTimeout"`
	WorkerNum   int    `yaml:"workerNum"`
	MaxConns    int    `yaml:"maxConns"`
	MaxIdle     int    `yaml:"maxIdle"`
	HbsMod      string `yaml:"hbsMod"`
	EventPrefix string `yaml:"eventPrefix"`
}

var (
	// config
	JudgeConfig JudgeSection
	Ident       string

	// 连接池 node_address -> connection_pool
	JudgeConnPools *pools.ConnPools

	// queue
	JudgeQueues = cache.SafeJudgeQueue{}

	DefaultSendTaskSleepInterval = time.Millisecond * 50 //默认睡眠间隔为50ms
	DefaultSendQueueMaxSize      = 102400
	MaxSendRetry                 = 10
)

func InitJudge(section JudgeSection, ident string) {
	JudgeConfig = section
	Ident = ident

	judges := GetJudges()

	// init connPool
	JudgeConnPools = pools.NewConnPools(JudgeConfig.MaxConns, JudgeConfig.MaxIdle, JudgeConfig.ConnTimeout, JudgeConfig.CallTimeout, judges)

	// init queue
	JudgeQueues = cache.NewJudgeQueue()
	for _, judgeNode := range judges {
		JudgeQueues.Set(judgeNode, list.NewSafeListLimited(DefaultSendQueueMaxSize))
	}

	// start task
	judgeConcurrent := JudgeConfig.WorkerNum
	if judgeConcurrent < 1 {
		judgeConcurrent = 1
	}
	judgeQueue := JudgeQueues.GetAll()
	for instance, queue := range judgeQueue {
		go Send2JudgeTask(queue, instance, judgeConcurrent)
	}

}

func Send2JudgeTask(Q *list.SafeListLimited, addr string, concurrent int) {
	batch := JudgeConfig.Batch
	sema := semaphore.NewSemaphore(concurrent)

	for {
		items := Q.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}
		judgeItems := make([]*dataobj.JudgeItem, count)
		stats.Counter.Set("points.out.judge", count)
		for i := 0; i < count; i++ {
			judgeItems[i] = items[i].(*dataobj.JudgeItem)
			logger.Debug("send to judge: ", judgeItems[i])
		}

		sema.Acquire()
		go func(addr string, judgeItems []*dataobj.JudgeItem, count int) {
			defer sema.Release()
			if strings.Contains(addr, Ident) {
				logger.Debugf("send judgeItems addr: %s,Ident:%s", addr, Ident)
				Send(judgeItems)
				return
			}
			resp := &dataobj.SimpleRpcResponse{}
			var err error
			sendOk := false
			for i := 0; i < MaxSendRetry; i++ {
				err = JudgeConnPools.Call(addr, "Server.Send", judgeItems, resp)
				if err == nil {
					sendOk = true
					break
				}
				logger.Warningf("send judge %s fail: %v", addr, err)
				time.Sleep(time.Millisecond * 10)
			}

			if !sendOk {
				stats.Counter.Set("points.out.err", count)
				for _, item := range judgeItems {
					logger.Errorf("send %v to judge %s fail: %v", item, addr, err)
				}
			}

		}(addr, judgeItems, count)
	}
}

func Push2JudgeQueue(items []*dataobj.MetricValue) {
	errCnt := 0
	for _, item := range items {
		var key string
		if item.Nid != "" {
			key = str.ToMD5(item.Nid, item.Metric, "")
		} else {
			key = str.ToMD5(item.Endpoint, item.Metric, "")
		}
		stras := cache.StraMap.GetByKey(key)

		for _, stra := range stras {
			if !TagMatch(stra.Tags, item.TagsMap) {
				continue
			}
			judgeItem := &dataobj.JudgeItem{
				Nid:       item.Nid,
				Endpoint:  item.Endpoint,
				Metric:    item.Metric,
				Value:     item.Value,
				Timestamp: item.Timestamp,
				DsType:    item.CounterType,
				Tags:      item.Tags,
				TagsMap:   item.TagsMap,
				Step:      int(item.Step),
				Sid:       stra.Id,
				Extra:     item.Extra,
			}

			q, exists := JudgeQueues.Get(stra.JudgeInstance)
			if exists {
				if !q.PushFront(judgeItem) {
					errCnt += 1
				}
			}
		}
	}
	stats.Counter.Set("judgeConfig.queue.err", errCnt)
}

func alignTs(ts int64, period int64) int64 {
	return ts - ts%period
}

func TagMatch(straTags []models.Tag, tag map[string]string) bool {
	for _, stag := range straTags {
		if _, exists := tag[stag.Tkey]; !exists {
			return false
		}
		var match bool
		if stag.Topt == "=" { //当前策略 tagkey 对应的 tagv
			for _, v := range stag.Tval {
				if tag[stag.Tkey] == v {
					match = true
					break
				}
			}
		} else {
			match = true
			for _, v := range stag.Tval {
				if tag[stag.Tkey] == v {
					match = false
					return match
				}
			}
		}

		if !match {
			return false
		}
	}
	return true
}

func GetJudges() []string {
	var judgeInstances []string
	instances, err := models.GetAllInstances("server", 1)
	if err != nil {
		stats.Counter.Set("server.get.err", 1)
		return judgeInstances
	}
	for _, instance := range instances {
		judgeInstance := instance.Identity + ":" + instance.RPCPort
		judgeInstances = append(judgeInstances, judgeInstance)
	}
	return judgeInstances
}
