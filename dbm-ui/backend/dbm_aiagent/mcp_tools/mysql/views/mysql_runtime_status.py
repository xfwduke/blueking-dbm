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
from rest_framework.response import Response
from django.utils.translation import gettext as _

from backend.components import DRSApi
from backend.dbm_aiagent.mcp_tools.constants import DBMMCPTags, DBMAMcpTools
from backend.dbm_aiagent.mcp_tools.decorators import mcp_tools_api_decorator
from backend.dbm_aiagent.mcp_tools.mysql.serializers.mysql_runtime_status import ShowProcesslistResponseSerializer, \
    MySQLInstanceAddressesSerializer
from backend.dbm_aiagent.mcp_tools.views import McpToolsViewSet


class MySQLRuntimeStatusMcpToolsViewSet(McpToolsViewSet):
    default_permission_class = []

    @mcp_tools_api_decorator(
        description=_("在 MySQL 实例上查询各自的当前时间"),
        request_slz=MySQLInstanceAddressesSerializer,
        response_slz=ShowProcesslistResponseSerializer,
        tags=[DBMMCPTags.READ],
        mcp=[DBMAMcpTools.DBM],
    )
    def select_now(self, request, *args, **kwargs):
        addresses = self.params_validate(self.get_serializer_class())["addresses"]
        drs_response = DRSApi.rpc(
            {
                "addresses": addresses,
                "cmds": ["select now() as current_time"],
                "force": True,
                "bk_cloud_id": 0,
                "query_timeout": 2
            }
        )

        res = []
        for row in drs_response:
            res.append(
                {
                    "address": row["address"],
                    "current_time": row["cmd_results"][0]["table_data"][0]["current_time"]
                }
            )

        return Response(res)
