package v2

import (
	"dbm-services/common/go-pubpkg/errno"
	"dbm-services/mysql/priv-service/handler"
	"dbm-services/mysql/priv-service/service/v2/drop_job_temp_account"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func DropJobTempAccount(c *gin.Context) {
	slog.Info("DropJobTempAccount")
	var input drop_job_temp_account.DropJobTempAccount

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

	err = input.DropUser()
	handler.SendResponse(c, err, nil)
	return
}
