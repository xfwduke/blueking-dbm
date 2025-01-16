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
from typing import List, Optional

from django.db.models import Q, QuerySet

from backend.components import DBConfigApi
from backend.db_meta.enums import MachineType
from backend.db_meta.models import Machine, ProxyInstance, StorageInstance
from backend.flow.consts import SYSTEM_DBS


def list_instance_monitor_config(bk_cloud_id: int, ip: str, port_list: Optional[List[int]] = None) -> List[dict]:
    m = Machine.objects.get(bk_cloud_id=bk_cloud_id, ip=ip)

    q = Q()
    q |= Q(**{"machine": m})

    if port_list:
        q &= Q(**{"port__in": port_list})

    res = []
    if m.machine_type in [MachineType.PROXY, MachineType.SPIDER]:
        qs = ProxyInstance.objects.filter(q).prefetch_related("cluster", "machine")
        res = generate_from_qs(bk_cloud_id=bk_cloud_id, qs=qs, has_role=False)
    elif m.machine_type in [MachineType.BACKEND, MachineType.SINGLE, MachineType.REMOTE]:
        qs = StorageInstance.objects.filter(q).prefetch_related("cluster", "machine")
        res = generate_from_qs(bk_cloud_id=bk_cloud_id, qs=qs, has_role=True)

    # zip_str = zlib.compress(json.dumps(res).encode("utf-8"))
    # print(len(zip_str), len(json.dumps(res).encode("utf-8")))

    return res


def generate_from_qs(bk_cloud_id: int, qs: QuerySet, has_role: bool) -> List[dict]:
    res = []
    for i in qs.all():
        cluster = i.cluster.all()[0]
        role = ""
        if has_role:
            role = i.instance_inner_role

        res.append(
            {
                "system_dbs": SYSTEM_DBS,
                "api_urs": "http://127.0.0.1:9999",
                "machine_type": i.machine_type,
                "bk_cloud_id": bk_cloud_id,
                "bk_biz_id": i.bk_biz_id,
                "ip": i.machine.ip,
                "port": i.port,
                "role": role,
                "bk_instance_id": i.bk_instance_id,
                "immute_domain": cluster.immute_domain,
                "db_module_id": cluster.db_module_id,
                "cluster_id": cluster.id,
                "items_config": DBConfigApi.query_conf_item(
                    {
                        "bk_biz_id": f"{cluster.bk_cloud_id}",
                        "level_name": "cluster",
                        "level_value": cluster.immute_domain,
                        "conf_file": "items-config.yaml",
                        "conf_type": "mysql_monitor",
                        "namespace": cluster.cluster_type,
                        "level_info": {"module": f"{cluster.db_module_id}"},
                        "format": "map",
                    }
                )["content"],
            }
        )

    return res
