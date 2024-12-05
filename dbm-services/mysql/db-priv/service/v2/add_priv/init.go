package add_priv

import (
	"dbm-services/mysql/priv-service/service"
	"encoding/json"
	"time"
)

type PrivTaskPara struct {
	*service.PrivTaskPara
}

func (c *PrivTaskPara) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

type TbAccountRules struct {
	Id          int64     `gorm:"column:id;primary_key;auto_increment" json:"id"`
	BkBizId     int64     `gorm:"column:bk_biz_id;not_null" json:"bk_biz_id"`
	ClusterType string    `gorm:"column:cluster_type;not_null" json:"cluster_type"`
	AccountId   int64     `gorm:"column:account_id;not_null" json:"account_id"`
	Dbname      string    `gorm:"column:dbname;not_null" json:"dbname"`
	Priv        string    `gorm:"column:priv;not_null" json:"priv"`
	DmlDdlPriv  string    `gorm:"column:dml_ddl_priv;not_null" json:"dml_ddl_priv"`
	GlobalPriv  string    `gorm:"column:global_priv;not_null" json:"global_priv"`
	Creator     string    `gorm:"column:creator;not_null;" json:"creator"`
	CreateTime  time.Time `gorm:"column:create_time" json:"create_time"`
	Operator    string    `gorm:"column:operator" json:"operator"`
	UpdateTime  time.Time `gorm:"column:update_time" json:"update_time"`
}
