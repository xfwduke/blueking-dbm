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
from collections import defaultdict
from typing import List, Tuple

from backend.db_meta.models import Cluster


def group_ips(cluster_objects: List[Cluster]) -> Tuple:
    """
    聚合集群 ip
    proxy_group: {
      bk_cloud_id(int): {
          ip(str): [ports]
        }
    }
    storage_group: {
      bk_cloud_id(int): {
          ip(str): [ports]
        }
    }
    """

    proxy_group = defaultdict(lambda: defaultdict(list))
    storage_group = defaultdict(lambda: defaultdict(list))

    for cluster in cluster_objects:
        cluster.proxyinstance_set.values("machine__ip")
        for i in cluster.proxyinstance_set.all():
            ip = i.machine.ip
            bk_cloud_id = i.machine.bk_cloud_id
            proxy_group[bk_cloud_id][ip].append(i.port)
        for i in cluster.storageinstance_set.all():
            ip = i.machine.ip
            bk_cloud_id = i.machine.bk_cloud_id
            storage_group[bk_cloud_id][ip].append(i.port)

    return proxy_group, storage_group
