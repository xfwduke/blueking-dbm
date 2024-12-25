package handler_rpc

import (
	redis_rpc2 "dbm-services/mysql/db-remote-service/pkg/rpc_implement/redis_rpc"
)

// RedisRPCHandler TODO
var RedisRPCHandler = redis_rpc2.NewRedisRPCEmbed().DoCommand

// TwemproxyRPCHandler TODO
var TwemproxyRPCHandler = redis_rpc2.NewTwemproxyRPCEmbed().DoCommand
