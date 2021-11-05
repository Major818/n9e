package all

import (
	// remote
	// _ "github.com/Major818/n9e/v4/src/modules/server/plugins/api"
	// telegraf style
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/dns_query"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/elasticsearch"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/github"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/haproxy"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/http_response"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/mongodb"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/mysql"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/net_response"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/nginx"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/ping"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/prometheus"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/rabbitmq"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/redis"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/tengine"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/zookeeper"

	// local
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/log"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/plugin"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/port"
	_ "github.com/Major818/n9e/v4/src/modules/server/plugins/proc"
)
