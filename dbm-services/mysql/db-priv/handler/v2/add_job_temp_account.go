package v2

import (
	"dbm-services/common/go-pubpkg/errno"
	"dbm-services/mysql/priv-service/handler"
	"dbm-services/mysql/priv-service/service/v2/add_job_temp_account"
	"encoding/json"
	"io"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
)

func AddJobTempAccount(c *gin.Context) {
	slog.Info("do AddJobTempAccount v2!")
	var input add_job_temp_account.AddJobTempAccount
	ticket := strings.ToUpper(c.DefaultQuery("ticket", "/priv/v2"))

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Error("msg", err)
		handler.SendResponse(c, errno.ErrBind, err)
		return
	}

	if err = json.Unmarshal(body, &input); err != nil {
		slog.Error("msg", err)
		handler.SendResponse(c, errno.ErrBind, err)
		return
	}

	err = input.AddPriv(string(body), ticket)
	handler.SendResponse(c, err, nil)
	return
}
