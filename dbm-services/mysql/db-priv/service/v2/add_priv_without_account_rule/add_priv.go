package add_priv_without_account_rule

import (
	"dbm-services/mysql/priv-service/service/v2/internal/drs"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/pkg/errors"
)

func (c *AddPrivWithoutAccountRule) AddPriv(jsonPara string, ticket string) error {
	drsRes, err := drs.RPCMySQL(
		c.BkCloudId,
		c.Addresses,
		[]string{
			fmt.Sprintf(
				`CALL infodba_schema.dba_grant_job_temp_account('%s', '%s')`,
				c.User, c.Psw,
			),
		},
		true,
		30,
	)
	if err != nil {
		slog.Error("add_priv_without_account_rule", slog.String("err", err.Error()))
		return err
	}
	slog.Info(
		"add_priv_without_account_rule",
		slog.String("response", fmt.Sprintf("%+v", drsRes)),
	)

	reports := make(map[string][]string)

	for _, r := range drsRes {
		if r.ErrorMsg != "" {
			err := errors.New(r.ErrorMsg)
			slog.Error(
				"add_priv_without_account_rule",
				slog.String("err", err.Error()),
				slog.String("addr", r.Address),
			)
			reports[r.Address] = []string{r.ErrorMsg}
			continue
		}
		if r.CmdResults[0].ErrorMsg != "" {
			if _, ok := reports[r.Address]; !ok {
				reports[r.Address] = make([]string, 0)
			}

			err := errors.New(r.CmdResults[0].ErrorMsg)
			slog.Error(
				"add_priv_without_account_rule",
				slog.String("err", err.Error()),
				slog.String("addr", r.Address),
			)
			reports[r.Address] = append(reports[r.Address], r.CmdResults[0].ErrorMsg)
		}
	}

	if len(reports) > 0 {
		slog.Info("add_priv_without_account_rule", slog.Any("reports", reports))
		b, err := json.Marshal(reports)
		if err != nil {
			slog.Error(
				"add_priv_without_account_rule",
				slog.String("err", err.Error()),
				slog.String("reports", string(b)),
			)
			return err
		}
		return errors.New(string(b))
	}

	slog.Info("add_priv_without_account_rule finish")
	return nil
}
