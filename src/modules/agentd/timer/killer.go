package timer

import (
	"fmt"
	"strings"

	"github.com/Major818/n9e/v4/src/modules/agentd/config"

	"github.com/toolkits/pkg/sys"
)

func KillProcessByTaskID(id int64) error {
	dir := strings.TrimRight(config.Config.Job.MetaDir, "/")
	arr := strings.Split(dir, "/")
	lst := arr[len(arr)-1]
	return sys.KillProcessByCmdline(fmt.Sprintf("%s/%d/script", lst, id))
}
