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
from backend.configuration.constants import DBType
from backend.db_services.mysql.permission.constants import CloneClusterType
from backend.ticket.constants import TicketType

BASE_FLOW_PARAMS = {
    "uid": 1,
    "root_id": "4146d614abf011",
    "ticket_type": "",
    "created_by": "admin",
    "bk_biz_id": 1,
}

MYSQL_AUTHORIZE_FLOW_PARAMS = {
    **BASE_FLOW_PARAMS,
    "db_type": DBType.MySQL.value,
    "ticket_type": TicketType.MYSQL_AUTHORIZE_RULES.value,
    "operator": "admin",
    "authorize_data": {
        "bk_biz_id": 1,
        "operator": "admin",
        "user": "admin",
        "access_dbs": ["datamain"],
        "account_rules": [{"dbname": "datamain", "bk_biz_id": 1}],
        "source_ips": ["127.0.0.1", "127.0.0.2"],
        "target_instances": ["gamedb.privtest55.blueking.db"],
        "cluster_type": "tendbha",
    },
}

MYSQL_CLONE_FLOW_PARAMS = {
    **BASE_FLOW_PARAMS,
    "operator": "admin",
    "clone_cluster_type": CloneClusterType.MYSQL,
    "ticket_type": TicketType.MYSQL_CLIENT_CLONE_RULES.value,
    "clone_data": {"source": "127.0.0.1:", "target": "127.0.1.1", "bk_cloud_id": 0},
    "clone_type": "client",
    "inst_machine_type_map": {},
}
