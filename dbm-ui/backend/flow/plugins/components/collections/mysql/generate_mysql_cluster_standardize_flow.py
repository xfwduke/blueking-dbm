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
from django.utils.translation import gettext as _
from pipeline.component_framework.component import Component

from backend.db_meta.models import Cluster
from backend.flow.engine.bamboo.scene.mysql.deploy_peripheraltools.departs import ALLDEPARTS
from backend.flow.plugins.components.collections.common.base_service import BaseService
from backend.ticket.constants import TicketType
from backend.ticket.models import Ticket


class GenerateMySQLClusterStandardizeFlowService(BaseService):
    def _execute(self, data, parent_data) -> bool:
        global_data = data.get_one_of_inputs("global_data")
        kwargs = data.get_one_of_inputs("kwargs")

        getattr(self, kwargs.get("trans_func"))(global_data, kwargs)
        return True

    @staticmethod
    def generate_from_immute_domains(global_data, kwargs):
        immute_domains = kwargs.get("immute_domains")
        cluster_objects = Cluster.objects.filter(immute_domain__in=immute_domains)

        ticket = Ticket.objects.get(id=global_data["uid"])
        bk_biz_id = global_data["bk_biz_id"]

        Ticket.create_ticket(
            ticket_type=TicketType.MYSQL_CLUSTER_STANDARDIZE,
            creator=global_data["created_by"],
            bk_biz_id=bk_biz_id,
            remark=_("集群标准化, 关联单据: {}".format(ticket.url)),
            details={
                "bk_biz_id": bk_biz_id,
                "cluster_type": cluster_objects.first().cluster_type,
                "cluster_ids": list(cluster_objects.values_list("id", flat=True)),
                "departs": kwargs.get("departs", ALLDEPARTS),
                "with_deploy_binary": kwargs.get("with_deploy_binary", True),
                "with_push_config": kwargs.get("with_push_config", True),
                "with_collect_sysinfo": kwargs.get("with_collect_sysinfo", True),
            },
        )


class GenerateMySQLClusterStandardizeFlowComponent(Component):
    name = __name__
    code = "generate_mysql_cluster_standardize_flow"
    bound_service = GenerateMySQLClusterStandardizeFlowService
