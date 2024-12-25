package handler_rpc

import (
	"dbm-services/mysql/db-remote-service/pkg/rpc_implement/webconsole_rpc"
)

var WebConsoleRPCHandler = generalHandler(&webconsole_rpc.WebConsoleRPC{})
