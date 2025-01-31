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

	"dbm-services/common/db-resource/internal/svr/bk"
	"dbm-services/common/go-pubpkg/logger"
)

// TbRpDetailArchive 资源池资源归档表
// nolint
type TbRpDetailArchive struct {
	ID              int                      `gorm:"primary_key;auto_increment;not_null" json:"-"`
	BkCloudID       int                      `gorm:"column:bk_cloud_id;type:int(11);not null;comment:'云区域 ID'" json:"bk_cloud_id"`
	BkBizId         int                      `gorm:"column:bk_biz_id;type:int(11);not null;comment:'机器当前所属业务'" json:"bk_biz_id"`
	DedicatedBiz    int                      `gorm:"column:dedicated_biz;type:int(11);default:0;comment:专属业务" json:"dedicated_biz"`
	RsType          string                   `gorm:"column:rs_type;type:varchar(64);default:'PUBLIC';comment:资源专用组件类型" json:"rs_type"`
	Bizs            map[string]string        `gorm:"-"`
	BkHostID        int                      `gorm:"column:bk_host_id;type:int(11);not null;comment:'bk主机ID'" json:"bk_host_id"`
	IP              string                   `gorm:"column:ip;type:varchar(20);not null" json:"ip"` //  svr ip
	AssetID         string                   `gorm:"column:asset_id;type:varchar(64);not null;comment:'固定资产编号'" json:"asset_id"`
	DeviceClass     string                   `gorm:"column:device_class;type:varchar(64);not null" json:"device_class"` //  对应机型 A30,D3
	SvrTypeName     string                   `gorm:"column:svr_type_name;type:varchar(64);not null;comment:'服务器型号,判断是否是云机器'" json:"svr_type_name"`
	CPUNum          int                      `gorm:"column:cpu_num;type:int(11);not null;comment:'cpu核数'" json:"cpu_num"`
	DramCap         int                      `gorm:"column:dram_cap;type:int(11);not null;comment:'内存大小'" json:"dram_cap"`
	StorageDevice   json.RawMessage          `gorm:"column:storage_device;type:json;comment:'磁盘设备'" json:"storage_device"`
	TotalStorageCap int                      `gorm:"column:total_storage_cap;type:int(11);comment:'磁盘总容量'" json:"total_storage_cap"`
	Storages        map[string]bk.DiskDetail `gorm:"-"`
	OsType          string                   `gorm:"column:os_type;type:varchar(32);not null;comment:'操作系统类型'" json:"os_type"`
	OsBit           string                   `gorm:"column:os_bit;type:varchar(32);not null;comment:'操作系统位数'" json:"os_bit"`
	//  操作系统版本
	OsVerion string `gorm:"column:os_version;type:varchar(64);not null;comment:'操作系统版本'" json:"os_version"`
	//  操作系统名称
	OsName string `gorm:"column:os_name;type:varchar(64);not null;comment:'操作系统名称'" json:"os_name"`
	// 磁盘Raid
	Raid string `gorm:"column:raid;type:varchar(20);not null" json:"raid"`
	//  实际城市ID
	CityID string `gorm:"column:city_id;type:varchar(64);not null" json:"city_id"`
	// 实际城市名称
	City string `gorm:"column:city;type:varchar(128);not null" json:"city"`
	//  园区, 例如光明 cc_device_szone
	SubZone string `gorm:"column:sub_zone;type:varchar(32);not null" json:"sub_zone"`
	//  园区ID cc_device_szone_id
	SubZoneID string `gorm:"column:sub_zone_id;type:varchar(64);not null" json:"sub_zone_id"`
	//  存放机架ID,判断是否是同机架
	RackID string `gorm:"column:rack_id;type:varchar(64);not null" json:"rack_id"`
	//  网络设备ID, 判断是同交换机
	NetDeviceID string `gorm:"column:net_device_id;type:varchar(128)" json:"net_device_id"`
	// 标签
	Labels string `gorm:"column:labels;type:json" json:"labels"`
	IsInit int    `gorm:"column:is_init;type:int(11);comment:'是否初始化过'" json:"-"`
	IsIdle int    `gorm:"column:is_idle;type:int(11);comment:'是否空闲检查过'" json:"-"`
	// Status: Unused: 未使用 Used: 已经售卖被使用: Preselected:预占用
	Status    string `gorm:"column:status;type:varchar(20);not null" json:"status"`
	BkAgentId string `gorm:"index:idx_bk_agent_id;column:bk_agent_id;type:varchar(64);not null" json:"bk_agent_id"`
	// gse Agent当前运行状态码, -1:未知 0:初始安装 1:启动中 2:运行中 3:有损状态 4:繁忙状态 5:升级中 6:停止中 7:解除安装
	AgentStatusCode int `gorm:"column:gse_agent_status_code;type:int(11);not null" json:"gse_agent_status_code"`
	// agent status 最后一次更新时间
	AgentStatusUpdateTime time.Time `gorm:"column:agent_status_update_time;type:timestamp;default:1970-01-01 08:00:01" json:"agent_status_update_time"`
	ConsumeTime           time.Time `gorm:"column:consume_time;type:timestamp;default:1970-01-01 08:00:01" json:"consume_time"`
	UpdateTime            time.Time `gorm:"column:update_time;type:timestamp;default:CURRENT_TIMESTAMP()" json:"update_time"`
	CreateTime            time.Time `gorm:"column:create_time;type:timestamp;default:CURRENT_TIMESTAMP()" json:"create_time"`
}

// initarchive 启动的时候归档未清理的资源
func initarchive() {
	tx := DB.Self.Begin()
	if err := tx.Exec("insert into tb_rp_detail_archive select * from tb_rp_detail where status = ? ", Used).
		Error; err != nil {
		logger.Error("insert into tb_rp_detail_archive failed %s", err.Error())
	}
	if err := tx.Exec("delete from tb_rp_detail where status = ? ", Used).Error; err != nil {
		logger.Error("delte from tb_rp_detail failed %s", err.Error())
	}
	tx.Commit()
}

// TableName table name
func (TbRpDetailArchive) TableName() string {
	return TbRpDetailArchiveName()
}

// TbRpDetailArchiveName  table name
func TbRpDetailArchiveName() string {
	return "tb_rp_detail_archive"
}

// ArchiverResouce 将申请完的资源转移到归档表
func ArchiverResouce(ids []int) (err error) {
	tx := DB.Self.Begin()
	defer func() {
		if err != nil {
			if tx.Rollback().Error != nil {
				logger.Error("archive resource exception %s,rollback failed!!", err)
			}
		}
	}()
	if err = tx.Exec("insert into tb_rp_detail_archive select * from tb_rp_detail where id in ? and status = ? ", ids,
		Used).Error; err != nil {
		return err
	}
	if err = tx.Exec("delete from tb_rp_detail where  id in ?  and status = ? ", ids, Used).Error; err != nil {
		return err
	}
	if err = tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
