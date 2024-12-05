package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Routes() []*gin.RouteInfo {
	return []*gin.RouteInfo{
		{Method: http.MethodPost, Path: "add_priv", HandlerFunc: AddPriv},
	}
}
