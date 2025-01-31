/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-DB管理系统(BlueKing-BK-DBM) available.
 * Copyright (C) 2017-2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package model

import (
	"encoding/json"
	"time"
)

const (
	// Consumed 消费主机
	Consumed = "consumed"
	// Imported 导入主机
	Imported = "imported"
)

const (
	// StatusSuccess operator success
	StatusSuccess = "success"
	// StatusFailed failed
	StatusFailed = "failed"
)

// TbRpOperationInfo 资源池操作记录表
// nolint
type TbRpOperationInfo struct {
	ID int `gorm:"primaryKey;auto_increment;not null" json:"-"`
	//nolint
	RequestID     string          `gorm:"index:idx_request_id;column:request_id;type:varchar(64);not null" json:"request_id"`
	TotalCount    int             `gorm:"column:total_count;type:int(11);comment:task Id" json:"total_count"`
	BkHostIds     json.RawMessage `gorm:"column:bk_host_ids;type:json;comment:主机Id" json:"bk_host_ids"`
	IpList        json.RawMessage `gorm:"column:ip_list;type:json;comment:主机ip" json:"ip_list"`
	OperationType string          `gorm:"column:operation_type;type:varchar(64);not null;comment:'operation type'" json:"operation_type"`
	Operator      string          `gorm:"column:operator;type:varchar(64);not null;comment:'operator user'" json:"operator"`
	Status        string          `gorm:"column:status;type:varchar(64);not null;comment: status" json:"-"`
	TaskId        string          `gorm:"column:task_id;type:varchar(128);not null;comment:'task Id'" json:"task_id"`
	BillId        string          `gorm:"column:bill_id;type:varchar(128);not null;comment:'bill Id'" json:"bill_id"`
	BillType      string          `gorm:"column:bill_type;type:varchar(128);not null;comment:'bill type'" json:"bill_type"`
	Description   string          `gorm:"column:description;type:varchar(256);default:'';comment:'description'" json:"description"`
	UpdateTime    time.Time       `gorm:"column:update_time;type:timestamp" json:"update_time"` // 最后修改时间
	CreateTime    time.Time       `gorm:"column:create_time;type:datetime" json:"create_time"`  // 创建时间
}

// TableName table name
func (TbRpOperationInfo) TableName() string {
	return TbRpOperationInfoTableName()
}

// TbRpOperationInfoTableName table name
func TbRpOperationInfoTableName() string {
	return "tb_rp_operation_info"
}

// getTbRpOperationInfoColumns 获取tb_rp_operation_info的字段名称
func getTbRpOperationInfoColumns() ([]string, error) {
	result, err := DB.Self.Migrator().ColumnTypes(&TbRpOperationInfo{})
	if err != nil {
		return []string{}, err
	}
	columns := []string{}
	for _, v := range result {
		columns = append(columns, v.Name())
	}
	// add Reverse sorting colums
	for _, v := range result {
		columns = append(columns, "-"+v.Name())
	}
	return columns, nil
}
