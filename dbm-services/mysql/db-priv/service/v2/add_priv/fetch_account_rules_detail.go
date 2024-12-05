package add_priv

import (
	"dbm-services/mysql/priv-service/service"
	"encoding/json"
	"log/slog"
)

/*
PrivTaskPara.TargetInstances 包含了所有目标域名
所有信息都应该基于这个预先获取
1. 获取目标域名的集群信息, 访问 db meta
2. 获取申请账号在目标实例上的已有信息, 访问 drs
3. 获取目标实例的版本信息, 访问 drs
*/

type accountAndRule struct {
	TbAccount          *service.TbAccounts
	TbAccountRulesList []*service.TbAccountRules
}

func (c *accountAndRule) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

// 这个函数查询量很小, 不需要做并发控制
func (c *PrivTaskPara) fetchAccountRulesDetail() (res *accountAndRule, err error) {
	res = &accountAndRule{
		TbAccount:          nil,
		TbAccountRulesList: make([]*service.TbAccountRules, 0),
	}
	// 这个 AccontRules 是传入的参数
	// 实际有用的是里面的 dbname
	// 因为 GetAccountRuleInfo 要传入一个 dbname
	// 所以这里需要循环跑
	for _, ele := range c.AccoutRules {
		account, accountRule, err := service.GetAccountRuleInfo(c.BkBizId, c.ClusterType, c.User, ele.Dbname)
		if err != nil {
			slog.Error(
				"fetch account rule detail failed",
				slog.Int64("bk biz id", c.BkBizId),
				slog.String("cluster type", c.ClusterType),
				slog.String("username", c.User),
				slog.String("dbname", ele.Dbname),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
		slog.Info(
			"fetch account rule detail",
			slog.Int64("bk biz id", c.BkBizId),
			slog.String("cluster type", c.ClusterType),
			slog.String("username", c.User),
			slog.String("dbname", ele.Dbname),
			slog.Any("account", account),
			slog.Any("accountRule", accountRule),
		)

		res.TbAccount = &account
		res.TbAccountRulesList = append(res.TbAccountRulesList, &accountRule)
	}

	return res, nil
}
