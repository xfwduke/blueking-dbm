package add_priv

import (
	"dbm-services/common/go-pubpkg/errno"
	"dbm-services/mysql/priv-service/service"
	"dbm-services/mysql/priv-service/service/v2/internal"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func (c *PrivTaskPara) AddPriv(jsonPara, ticket string) (err error) {
	slog.Info(
		"add priv",
		slog.String("jsonPara", jsonPara),
	)

	if c.ClusterType == internal.ClusterTypeSqlServerHA ||
		c.ClusterType == internal.ClusterTypeSqlServer ||
		c.ClusterType == internal.ClusterTypeSqlServerSingle {
		return c.AddPrivForSqlserver(jsonPara)
	}

	/*
		BkBizId != 0
		ClusterType != ""
		客户端 IP 去重
		目标实例/集群 去重
		BkBizId, ClusterType, User Dbname 能查到规则明细
	*/
	if c.BkBizId == 0 {
		return errno.BkBizIdIsEmpty
	}
	if c.ClusterType == "" {
		return errno.ClusterTypeIsEmpty
	}

	c.SourceIPs = internal.UniqueStringSlice(c.SourceIPs)
	// targetInstance 传入的其实全是域名
	c.TargetInstances = internal.UniqueStringSlice(c.TargetInstances)

	slog.Info("add priv", slog.String("source ips", strings.Join(c.SourceIPs, ",")))

	// 写审计日志
	service.AddPrivLog(
		service.PrivLog{
			Id:       0,
			BkBizId:  c.BkBizId,
			Ticket:   ticket,
			Operator: c.Operator,
			Para:     jsonPara,
			Time:     time.Now(),
		})

	// 目标实例的 dbmeta 信息
	targetMetaInfos, err := c.fetchTargetDBMetaInfo()
	if err != nil {
		slog.Error("add priv", slog.String("err", err.Error()))
		return err
	}
	slog.Info("add priv", slog.Any("target meta infos", targetMetaInfos))

	/*
		TenDBSingle 授权是在存储实例操作
		TenDBHA, 主备存储实例都要操作, 但权限有差异; proxy 看情况
		TenDBCluster 全都在 spider, 角色有差异
	*/

	// 开白名单
	// proxy 白名单是前置集中开, 所有出错了直接返回
	if c.ClusterType == internal.ClusterTypeTenDBHA {
		err = c.addWhiteList(targetMetaInfos)
		if err != nil {
			slog.Error("add priv", slog.String("err", err.Error()))
			return err
		}
	}

	// 接下来可以说都是面对 mysql 实例的授权了
	// 需要注意的是, TenDBHA 有些时候需要把 client ip 替换成 proxy ip
	// 所以 TenDBSingle 和 TenDBCluster 的授权语句对于所有 mysql 实例肯定是一样的
	// TenDBHA 如果申请的是 slave 权限, 也是一样的
	// TenDBHA 如果申请的是 master 权限, 并且有 padding Proxy, 有一部分是一样的
	clientIps, workingMySQLInstances := c.prepareMySQLPayload(targetMetaInfos)
	slog.Info(
		"add priv",
		slog.String("clientIps", strings.Join(clientIps, ",")),
		slog.Any("workingMySQLInstances", workingMySQLInstances),
	)

	// 获取相关的账号规则详情
	// 这里面就包含了权限明细, dbname, 密码啥的
	accountAndRuleDetails, err := c.fetchAccountRulesDetail()
	if err != nil {
		slog.Error("add priv", slog.String("err", err.Error()))
		return err
	}
	slog.Info(
		"add priv",
		slog.String("accountAndRuleDetails", accountAndRuleDetails.String()),
	)

	// err 是调用函数出错, 直接报错返回
	// reports 是实施授权的报告
	reports, err := c.addOnMySQL(clientIps, workingMySQLInstances, accountAndRuleDetails)
	if err != nil {
		slog.Error("add priv", slog.String("err", err.Error()))
		return err
	}
	if len(reports) > 0 {
		slog.Info("add priv", slog.Any("reports", reports))
		b, err := json.Marshal(reports)
		if err != nil {
			slog.Error("add priv", slog.String("err", err.Error()), slog.String("reports", string(b)))
			return err
		}
		return errors.New(string(b))
	}

	slog.Info("add priv finish")
	return nil
}
