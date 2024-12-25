package service

import (
	"dbm-services/mysql/db-remote-service/pkg/service/handler_rpc"

	"github.com/gin-gonic/gin"
)

// RegisterRouter 服务路由
func RegisterRouter(engine *gin.Engine) {
	mysqlGroup := engine.Group("/mysql")
	mysqlGroup.POST("/rpc", handler_rpc.MySQLRPCHandler)
	// 复杂接口的设计和其他的都不一样, 所以只能直接添加 handler, 不太好复用 rpc interface
	mysqlGroup.POST("/complex-rpc", handler_rpc.MySQLComplexHandler)

	proxyGroup := engine.Group("/proxy-admin")
	proxyGroup.POST("/rpc", handler_rpc.ProxyRPCHandler)

	redisGroup := engine.Group("/redis")
	redisGroup.POST("/rpc", handler_rpc.RedisRPCHandler)

	twemproxyGroup := engine.Group("/twemproxy")
	twemproxyGroup.POST("/rpc", handler_rpc.TwemproxyRPCHandler)

	sqlserverGroup := engine.Group("/sqlserver")
	sqlserverGroup.POST("/rpc", handler_rpc.SqlserverRPCHandler)

	webConsoleGroup := engine.Group("/webconsole")
	webConsoleGroup.POST("/rpc", handler_rpc.WebConsoleRPCHandler)
}
