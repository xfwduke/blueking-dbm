package handler_rpc

import (
	"dbm-services/mysql/db-remote-service/pkg/rpc_implement/mysql_rpc"
)

// MySQLRPCHandler mysql 请求响应
var MySQLRPCHandler = generalHandler(&mysql_rpc.MySQLRPCEmbed{})
