/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-DB管理系统(BlueKing-BK-DBM) available.
 * Copyright (C) 2017-2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package manage

import (
	"fmt"
	"path"
	"strings"

	rf "github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"dbm-services/common/db-resource/internal/model"
	"dbm-services/common/db-resource/internal/svr/bk"
	"dbm-services/common/db-resource/internal/svr/dbmapi"
	"dbm-services/common/db-resource/internal/svr/meta"
	"dbm-services/common/go-pubpkg/cmutil"
	"dbm-services/common/go-pubpkg/errno"
	"dbm-services/common/go-pubpkg/logger"
)

// MachineResourceGetterInputParam TODO
type MachineResourceGetterInputParam struct {
	// 专用业务Ids
	ForBiz       int               `json:"for_biz"`
	City         []string          `json:"city"`
	SubZoneIds   []string          `json:"subzone_ids"`
	DeviceClass  []string          `json:"device_class"`
	Labels       []string          `json:"labels"`
	Hosts        []string          `json:"hosts"`
	BkCloudIds   []int             `json:"bk_cloud_ids"`
	RsType       string            `json:"resource_type"`
	MountPoint   string            `json:"mount_point"`
	Cpu          meta.MeasureRange `json:"cpu"`
	Mem          meta.MeasureRange `json:"mem"`
	Disk         meta.MeasureRange `json:"disk"`
	DiskType     string            `json:"disk_type"`
	OsType       string            `json:"os_type"`
	StorageSpecs []meta.DiskSpec   `json:"storage_spec"`
	// 适用于用户没选业务和db类型的情况
	SetBizEmpty    bool `json:"set_empty_biz"`
	SetRsTypeEmpty bool `json:"set_empty_resource_type"`
	// true,false,""
	GseAgentAlive string `json:"gse_agent_alive"`
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
}

// List TODO
func (c *MachineResourceHandler) List(r *rf.Context) {
	var input MachineResourceGetterInputParam
	var count int64

	if c.Prepare(r, &input) != nil {
		return
	}
	if err := input.paramCheck(); err != nil {
		c.SendResponse(r, errno.ErrErrInvalidParam.AddErr(err), nil)
		return
	}
	db := model.DB.Self.Table(model.TbRpDetailName())
	if err := input.queryBs(db); err != nil {
		c.SendResponse(r, err, err.Error())
		return
	}
	if err := db.Count(&count).Error; err != nil {
		c.SendResponse(r, err, err.Error())
		return
	}
	if input.Limit > 0 {
		db = db.Offset(input.Offset).Limit(input.Limit)
	}
	var data []model.TbRpDetail
	if err := db.Find(&data).Error; err != nil {
		c.SendResponse(r, errno.ErrDBQuery.AddErr(err), err.Error())
		return
	}
	c.SendResponse(r, nil, map[string]interface{}{"details": data, "count": count})
}

func (c *MachineResourceGetterInputParam) paramCheck() (err error) {
	if !c.Cpu.Iegal() {
		return fmt.Errorf("非法参数 cpu min:%d,max:%d", c.Cpu.Min, c.Cpu.Max)
	}
	if !c.Mem.Iegal() {
		return fmt.Errorf("非法参数 mem min:%d,max:%d", c.Mem.Min, c.Mem.Max)
	}
	return nil
}

// matchStorageSpecs 匹配磁盘
func (c *MachineResourceGetterInputParam) matchStorageSpecs(db *gorm.DB) {
	if len(c.StorageSpecs) > 0 {
		for _, d := range c.StorageSpecs {
			if cmutil.IsNotEmpty(d.MountPoint) {
				mp := path.Clean(d.MountPoint)
				if cmutil.IsNotEmpty(d.DiskType) {
					db.Where(model.JSONQuery("storage_device").Equals(d.DiskType, mp, "disk_type"))
				}
				logger.Info("storage spec is %v", d)
				switch {
				case d.MaxSize > 0:
					db.Where(model.JSONQuery("storage_device").NumRange(d.MinSize, d.MaxSize, mp, "size"))
				case d.MaxSize <= 0 && d.MinSize > 0:
					db.Where(model.JSONQuery("storage_device").Gte(d.MinSize, mp, "size"))
				}
			}
		}
	} else {
		c.Disk.MatchTotalStorageSize(db)
		if cmutil.IsNotEmpty(c.MountPoint) {
			mp := path.Clean(c.MountPoint)
			if cmutil.IsNotEmpty(c.DiskType) {
				db.Where(model.JSONQuery("storage_device").Equals(c.DiskType, mp, "disk_type"))
			} else {
				db.Where(model.JSONQuery("storage_device").KeysContains([]string{mp}))
			}
		} else if cmutil.IsNotEmpty(c.DiskType) {
			db.Where(model.JSONQuery("storage_device").SubValContains(c.DiskType, "disk_type"))
		}
	}
}

func (c *MachineResourceGetterInputParam) getRealCitys() (realCistys []string, err error) {
	for _, logicCity := range c.City {
		rcitys, err := dbmapi.GetIdcCityByLogicCity(logicCity)
		if err != nil {
			logger.Error("from %s get real citys failed %s", logicCity, err.Error())
			return nil, err
		}
		realCistys = append(realCistys, rcitys...)
	}
	logger.Info("get real citys %v", realCistys)
	return
}

func (c *MachineResourceGetterInputParam) matchSpec(db *gorm.DB) {
	if len(c.DeviceClass) > 0 {
		switch {
		case c.Cpu.IsEmpty() && c.Mem.IsEmpty():
			db.Where(" device_class in (?) ", c.DeviceClass)
		case c.Cpu.IsEmpty() && c.Mem.IsNotEmpty():
			db.Where("? or device_class in (?)", c.Mem.MatchMemBuilder(), c.DeviceClass)
		case c.Cpu.IsNotEmpty() && c.Mem.IsEmpty():
			db.Where("? or device_class in (?)", c.Cpu.MatchCpuBuilder(), c.DeviceClass)
		case c.Cpu.IsNotEmpty() && c.Mem.IsNotEmpty():
			db.Where("( ? and  ? ) or device_class in (?)", c.Cpu.MatchCpuBuilder(), c.Mem.MatchMemBuilder(), c.DeviceClass)
		}
		return
	}
	c.Cpu.MatchCpu(db)
	c.Mem.MatchMem(db)
}
func (c *MachineResourceGetterInputParam) queryBs(db *gorm.DB) (err error) {
	db.Where("status = ? ", model.Unused)
	if len(c.Hosts) > 0 {
		db.Where("ip in (?)", c.Hosts)
		return nil
	}
	switch strings.TrimSpace(strings.ToLower(c.GseAgentAlive)) {
	case "true":
		db.Where("gse_agent_status_code = ?  ", bk.GSE_AGENT_OK)
	case "false":
		db.Where("gse_agent_status_code != ?  ", bk.GSE_AGENT_OK)
	}
	if len(c.BkCloudIds) > 0 {
		db.Where("bk_cloud_id in (?) ", c.BkCloudIds)
	}
	if !c.SetRsTypeEmpty {
		db.Where("rs_type = ? ", c.RsType)
	}
	if !c.SetBizEmpty {
		db.Where("dedicated_biz = ?", c.ForBiz)
	}
	c.matchSpec(db)
	c.matchStorageSpecs(db)
	if len(c.City) > 0 {
		realCitys, err := c.getRealCitys()
		if err != nil {
			return err
		}
		db.Where(" city in (?) ", realCitys)
	}
	if len(c.SubZoneIds) > 0 {
		db.Where(" sub_zone_id in (?) ", c.SubZoneIds)
	}
	if len(c.Labels) > 0 {
		db.Where(model.JSONQuery("labels").JointOrContains(c.Labels))
	}
	if cmutil.IsNotEmpty(c.OsType) {
		db.Where("os_type = ?", c.OsType)
	}
	db.Order("create_time desc")
	return nil
}

// ListAll TODO
func (c *MachineResourceHandler) ListAll(r *rf.Context) {
	// requestId := r.GetString("request_id")
	var data []model.TbRpDetail
	db := model.DB.Self.Table(model.TbRpDetailName()).Where("status in (?)", []string{model.Unused, model.Prepoccupied,
		model.Preselected})
	err := db.Scan(&data).Error
	if err != nil {
		c.SendResponse(r, err, err.Error())
		return
	}
	var count int64
	if err := db.Count(&count).Error; err != nil {
		c.SendResponse(r, err, err.Error())
		return
	}
	c.SendResponse(r, nil, map[string]interface{}{"details": data, "count": count})
}
