package stra

import (
	"encoding/json"
	"time"

	"github.com/Major818/n9e/v4/src/common/client"
	"github.com/Major818/n9e/v4/src/models"
	"github.com/Major818/n9e/v4/src/modules/agentd/config"

	"github.com/toolkits/pkg/logger"
)

func GetCollects() {
	if !config.Config.Stra.Enable {
		return
	}

	go loopDetect()
}

func loopDetect() {
	for {
		detect()
		time.Sleep(time.Duration(config.Config.Stra.Interval) * time.Second)
	}
}

func detect() {
	var resp string
	var c models.Collect
	err := client.GetCli("server").Call("Server.GetCollectBy", config.Endpoint, &resp)
	if err != nil {
		logger.Error("get collects err:", err)
		return
	}

	err = json.Unmarshal([]byte(resp), &c)
	if err != nil {
		logger.Error("get collects %s unmarshal err:", resp, err)
		return
	}

	logger.Debugf("get collect:%+v", c)
	Collect.Update(&c)
}
