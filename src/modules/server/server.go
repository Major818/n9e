package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Major818/n9e/v4/src/common/i18n"
	"github.com/Major818/n9e/v4/src/common/identity"
	"github.com/Major818/n9e/v4/src/common/loggeri"
	"github.com/Major818/n9e/v4/src/common/stats"
	"github.com/Major818/n9e/v4/src/models"
	"github.com/Major818/n9e/v4/src/modules/server/aggr"
	"github.com/Major818/n9e/v4/src/modules/server/alarm"
	"github.com/Major818/n9e/v4/src/modules/server/auth"
	"github.com/Major818/n9e/v4/src/modules/server/backend"
	"github.com/Major818/n9e/v4/src/modules/server/cache"
	"github.com/Major818/n9e/v4/src/modules/server/collector"
	"github.com/Major818/n9e/v4/src/modules/server/config"
	"github.com/Major818/n9e/v4/src/modules/server/cron"
	"github.com/Major818/n9e/v4/src/modules/server/http"
	"github.com/Major818/n9e/v4/src/modules/server/http/session"
	"github.com/Major818/n9e/v4/src/modules/server/judge"
	"github.com/Major818/n9e/v4/src/modules/server/judge/query"
	"github.com/Major818/n9e/v4/src/modules/server/rabbitmq"
	"github.com/Major818/n9e/v4/src/modules/server/redisc"
	"github.com/Major818/n9e/v4/src/modules/server/rpc"
	"github.com/Major818/n9e/v4/src/modules/server/ssoc"
	"github.com/Major818/n9e/v4/src/modules/server/timer"
	"github.com/Major818/n9e/v4/src/modules/server/wechat"

	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/all"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/api"

	_ "github.com/go-sql-driver/mysql"
	pcache "github.com/toolkits/pkg/cache"
	"github.com/toolkits/pkg/logger"
)

func main() {
	parseConf()
	conf := config.Config

	loggeri.Init(conf.Logger)
	i18n.Init()
	pcache.InitMemoryCache(time.Hour)

	checkIdentity()

	// 初始化数据库和相关数据
	models.InitMySQL("rdb", "mon", "ams", "hbs")
	if config.Config.Nems.Enabled {
		models.InitMySQL("nems")
		go cron.SyncSnmpCollects()
		go cron.SyncHardwares()
	}

	if conf.Rdb.SSO.Enable && conf.Rdb.Auth.ExtraMode.Enable {
		models.InitMySQL("sso")
	}
	models.InitSalt()
	models.InitRooter()

	ssoc.InitSSO()

	// 初始化 redis 用来处理告警事件、发送邮件短信等
	redisc.InitRedis(conf.Redis)

	// 初始化 rabbitmq 处理部分异步逻辑
	wechat.Init(conf.WeChat)
	rabbitmq.Init(conf.RabbitMQ)
	session.Init()

	auth.Init(conf.Rdb.Auth.ExtraMode)
	auth.Start()

	models.InitLDAP(conf.Rdb.LDAP)
	go stats.Init("n9e")

	if conf.Job.Enable {
		models.InitMySQL("job")
		timer.CacheHostDoing()
		go timer.Heartbeat()
		go timer.Schedule()
		go timer.CleanLong()
	}

	aggr.Init(conf.Transfer.Aggr)
	backend.Init(conf.Transfer.Backend)
	// init judge
	go judge.InitJudge(conf.Judge.Backend, config.Ident)

	cache.Init(conf.Monapi.Region)
	cron.Init()
	go cron.InitWorker(conf.Rdb.Sender)
	go cron.InitReportHeartBeat(conf.Report)

	//judge
	go query.Init(conf.Judge.Query)
	go cron.GetJudgeStrategy(conf.Judge.Strategy)
	go judge.NodataJudge(conf.Judge.NodataConcurrency)

	if conf.Monapi.AlarmEnabled {
		if err := alarm.SyncMaskconf(); err != nil {
			log.Fatalf("sync maskconf fail: %v", err)
		}

		if err := alarm.SyncStra(); err != nil {
			log.Fatalf("sync stra fail: %v", err)
		}

		go alarm.SyncMaskconfLoop()
		go alarm.SyncStraLoop()
		go alarm.CleanStraLoop()
		go alarm.ReadHighEvent()
		go alarm.ReadLowEvent()
		go alarm.CallbackConsumer()
		go alarm.MergeEvent()
		go alarm.CleanEventLoop()
	}

	if conf.Monapi.ApiDetectorEnabled {
		go cron.CheckDetectorNodes()
		go cron.SyncApiCollects()
	}

	if conf.Monapi.SnmpDetectorEnabled {
		go cron.CheckSnmpDetectorNodes()
	}

	if conf.Transfer.Aggr.Enabled {
		go cron.SyncAggrCalcStras()
		go cron.GetAggrCalcStrategy()
	}

	pluginInfo()

	go rpc.Start()

	http.Start()

	endingProc()
}

func parseConf() {
	if err := config.Parse(); err != nil {
		fmt.Println("cannot parse configuration file:", err)
		os.Exit(1)
	}
}

func checkIdentity() {
	ip, err := identity.GetIP()
	if err != nil {
		fmt.Println("cannot get ip:", err)
		os.Exit(1)
	}

	if ip == "127.0.0.1" {
		fmt.Println("identity: 127.0.0.1, cannot work")
		os.Exit(2)
	}
}

func endingProc() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-c:
		fmt.Printf("stop signal caught, stopping... pid=%d\n", os.Getpid())
	}

	logger.Close()
	http.Shutdown()
	fmt.Println("process stopped successfully")
}

func pluginInfo() {
	fmt.Println("remote collector")
	for k, v := range collector.GetRemoteCollectors() {
		fmt.Printf("  %d %s\n", k, v)
	}
}
