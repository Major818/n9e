package rpc

import (
	"github.com/Major818/n9e/v4/src/common/dataobj"
	"github.com/Major818/n9e/v4/src/modules/server/backend"

	"github.com/toolkits/pkg/logger"
)

func (t *Server) Query(args []dataobj.QueryData, reply *dataobj.QueryDataResp) error {
	dataSource, err := backend.GetDataSourceFor("")
	if err != nil {
		logger.Warningf("could not find datasource")
		return err
	}
	reply.Data = dataSource.QueryData(args)
	return nil
}
