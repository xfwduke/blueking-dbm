package handler_rpc

import (
	"dbm-services/mysql/db-remote-service/pkg/rpc_implement/proxy_rpc"
)

// ProxyRPCHandler proxy 请求响应
var ProxyRPCHandler = generalHandler(&proxy_rpc.ProxyRPCEmbed{})
