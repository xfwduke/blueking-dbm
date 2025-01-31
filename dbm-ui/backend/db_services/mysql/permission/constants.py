# -*- coding: utf-8 -*-
"""
TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-DB管理系统(BlueKing-BK-DBM) available.
Copyright (C) 2017-2023 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at https://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""

from django.utils.translation import ugettext_lazy as _

from blue_krill.data_types.enum import EnumField, StructuredEnum


class CloneClusterType(str, StructuredEnum):
    """克隆的集群类型"""

    MYSQL = EnumField("mysql", _("MySQL"))
    TendbCluster = EnumField("tendbcluster", _("TendbCluster"))


class CloneType(str, StructuredEnum):
    """权限克隆类型"""

    INSTANCE = EnumField("instance", _("实例克隆"))
    CLIENT = EnumField("client", _("客户端克隆"))


class CloneExcelTypeID(str, StructuredEnum):
    """下载权限克隆excel提供的id类型"""

    TICKET_ID = EnumField("ticket_id", _("单据ID"))
    CLONE_UID = EnumField("clone_id", _("克隆数据缓存UID"))


# EXCEL模板路径
AUTHORIZE_EXCEL_ERROR_TEMPLATE = "backend/db_services/mysql/excel_files/authorize_err_tpl.xlsx"
# 授权表头信息
AUTHORIZE_EXCEL_HEADER = ["账号(单个)", "访问源(多个)", "访问集群域名(多个)", "访问DB名(多个)"]

# 颜色常量代码
GREEN_COLOR_CODE = "7CCD7C"
GREY_COLOR_CODE = "9C9C9C"
RED_COLOR_CODE = "EE2C2C"

CLONE_INSTANCE_EXCEL_HEADER = ["旧实例", "新实例", "云区域ID"]
CLONE_CLIENT_EXCEL_HEADER = ["旧客户端IP", "新客户端IP", "云区域ID"]
CLONE_CLIENT_EXCEL_ERROR_HEADER = CLONE_CLIENT_EXCEL_HEADER + ["错误信息/提示信息"]
CLONE_INSTANCE_EXCEL_ERROR_HEADER = CLONE_INSTANCE_EXCEL_HEADER + ["错误信息/提示信息"]
CLONE_INSTANCE_EXCEL_HEADER_STYLE = {
    "旧实例": GREY_COLOR_CODE,
    "新实例": GREEN_COLOR_CODE,
    "云区域ID": GREEN_COLOR_CODE,
    "错误信息/提示信息": RED_COLOR_CODE,
}
CLONE_CLIENT_EXCEL_HEADER_STYLE = {
    "旧客户端IP": GREY_COLOR_CODE,
    "新客户端IP": GREEN_COLOR_CODE,
    "云区域ID": GREEN_COLOR_CODE,
    "错误信息/提示信息": RED_COLOR_CODE,
}

CLONE_EXCEL_HEADER_MAP = {
    CloneType.INSTANCE.value: CLONE_INSTANCE_EXCEL_HEADER,
    CloneType.CLIENT.value: CLONE_CLIENT_EXCEL_HEADER,
}
CLONE_EXCEL_ERROR_HEADER_MAP = {
    CloneType.INSTANCE.value: CLONE_INSTANCE_EXCEL_ERROR_HEADER,
    CloneType.CLIENT.value: CLONE_CLIENT_EXCEL_ERROR_HEADER,
}
CLONE_EXCEL_STYLE_HEADER_MAP = {
    CloneType.INSTANCE.value: CLONE_INSTANCE_EXCEL_HEADER_STYLE,
    CloneType.CLIENT.value: CLONE_CLIENT_EXCEL_HEADER_STYLE,
}

# 缓存数据过期时间
AUTHORIZE_DATA_EXPIRE_TIME = 60 * 60 * 6
CLONE_DATA_EXPIRE_TIME = 60 * 60 * 6
