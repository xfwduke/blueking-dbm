package peripheraltools

import (
	"dbm-services/common/go-pubpkg/logger"
	"dbm-services/mysql/db-tools/dbactuator/pkg/components"
	"dbm-services/mysql/db-tools/dbactuator/pkg/components/peripheraltools/checksum"
	"dbm-services/mysql/db-tools/dbactuator/pkg/components/peripheraltools/crond"
	"dbm-services/mysql/db-tools/dbactuator/pkg/components/peripheraltools/dba_toolkit"
	"dbm-services/mysql/db-tools/dbactuator/pkg/components/peripheraltools/dbbackup"
	"dbm-services/mysql/db-tools/dbactuator/pkg/components/peripheraltools/monitor"
	"dbm-services/mysql/db-tools/dbactuator/pkg/components/peripheraltools/rotatebinlog"

	"github.com/pkg/errors"
)

const (
	DepartDBAToolKit         = "dba-toolkit"
	DepartMySQLCrond         = "mysql-crond"
	DepartMySQLMonitor       = "mysql-monitor"
	DepartMySQLDBBackup      = "mysql-dbbackup"
	DepartMySQLRotateBinlog  = "rotate-binlog"
	DepartMySQLTableChecksum = "mysql-checksum"
)

type PrepareBinary struct {
	GeneralParam *components.GeneralParam `json:"general"`
	Params       *PrepareBinaryParams     `json:"extend"`
}

type PrepareBinaryParams struct {
	/*
		key 是
		    MySQLCrond = EnumField("mysql-crond", _("mysql-rond"))
		    MySQLMonitor = EnumField("mysql-monitor", _("mysql-monitor"))
		    MySQLDBBackup = EnumField("mysql-dbbackup", _("mysql-dbbackup"))
		    MySQLRotateBinlog = EnumField("rotate-binlog", _("rotate-binlog"))
		    MySQLTableChecksum = EnumField("mysql-checksum", _("mysql-checksum"))
	*/
	Departs map[string]*components.Medium `json:"departs"`
}

func (c *PrepareBinary) Run() (err error) {
	for k, _ := range c.Params.Departs {
		err = c.prepareOne(k)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
	}
	return nil
}

func (c *PrepareBinary) prepareOne(depart string) (err error) {
	switch depart {
	case DepartMySQLCrond:
		return crond.DeployBinary(c.Params.Departs[DepartMySQLCrond])
	case DepartMySQLMonitor:
		return monitor.DeployBinary(c.Params.Departs[DepartMySQLMonitor])
	case DepartMySQLDBBackup:
		return dbbackup.DeployBinary(c.Params.Departs[DepartMySQLDBBackup])
	case DepartMySQLRotateBinlog:
		return rotatebinlog.DeployBinary(c.Params.Departs[DepartMySQLRotateBinlog])
	case DepartMySQLTableChecksum:
		return checksum.DeployBinary(c.Params.Departs[DepartMySQLTableChecksum])
	case DepartDBAToolKit:
		return dba_toolkit.DeployBinary(c.Params.Departs[DepartDBAToolKit])
	default:
		err = errors.New("unknown depart " + depart)
		logger.Error(err.Error())
		return err
	}
}

func (c *PrepareBinary) Example() interface{} {
	return PrepareBinaryParams{Departs: map[string]*components.Medium{
		DepartMySQLCrond: {
			Pkg:    "mysql-crond",
			PkgMd5: "12346",
		},
	}}
}
