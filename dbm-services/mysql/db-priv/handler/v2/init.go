package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Routes() []*gin.RouteInfo {
	return []*gin.RouteInfo{
		{Method: http.MethodPost, Path: "add_priv", HandlerFunc: AddPriv},
		{Method: http.MethodPost, Path: "add_priv_without_account_rule", HandlerFunc: AddPrivWithoutAccountRule},
		{Method: http.MethodPost, Path: "drop_job_temp_account", HandlerFunc: DropJobTempAccount},
	}
}
