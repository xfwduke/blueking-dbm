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
import logging

from django.db import transaction

from backend.db_meta.enums import (
    ClusterEntryRole,
    ClusterEntryType,
    InstanceInnerRole,
    InstancePhase,
    InstanceRoleInstanceInnerRoleMap,
    InstanceStatus,
)
from backend.db_meta.models import Cluster, StorageInstance
from backend.flow.utils.mysql.mysql_module_operate import MysqlCCTopoOperator

logger = logging.getLogger("root")


@transaction.atomic
def switch_storage(cluster_id: int, target_storage_ip: str, origin_storage_ip: str, role: str = None):
    """
    集群主从成对迁移切换场景的元数据写入
    """
    cluster = Cluster.objects.get(id=cluster_id)
    cluster_storage_port = StorageInstance.objects.filter(cluster=cluster).all()[0].port
    target_storage = StorageInstance.objects.get(
        machine__ip=target_storage_ip, port=cluster_storage_port, machine__bk_cloud_id=cluster.bk_cloud_id
    )
    origin_storage = StorageInstance.objects.get(
        machine__ip=origin_storage_ip, port=cluster_storage_port, machine__bk_cloud_id=cluster.bk_cloud_id
    )
    cluster.storageinstance_set.remove(origin_storage)
    target_storage.status = InstanceStatus.RUNNING.value
    target_storage.phase = InstancePhase.ONLINE.value
    # target实例需要继承source实例的is_standby特性
    target_storage.is_stand_by = origin_storage.is_stand_by
    if role:
        # 如果是REPEATER角色，改成传入的role变量
        target_storage.instance_role = role
        target_storage.instance_inner_role = InstanceRoleInstanceInnerRoleMap[role].value
        target_storage.save()
        # 更新cmdb标签状态
        cc_topo_operator = MysqlCCTopoOperator(cluster)
        cc_topo_operator.is_bk_module_created = True
        cc_topo_operator.transfer_instances_to_cluster_module(instances=[target_storage], is_increment=True)
    else:
        target_storage.save()
    origin_storage.status = InstanceStatus.UNAVAILABLE.value
    origin_storage.phase = InstancePhase.OFFLINE.value
    origin_storage.is_stand_by = False
    origin_storage.save()


def change_proxy_storage_entry(cluster_id: int, master_ip: str, new_master_ip: str):
    cluster = Cluster.objects.get(id=cluster_id)
    master_storage = cluster.storageinstance_set.get(machine__ip=master_ip)
    new_master_storage = cluster.storageinstance_set.get(machine__ip=new_master_ip)
    proxy_list = master_storage.proxyinstance_set.all()
    for proxy in proxy_list:
        proxy.storageinstance.remove(master_storage)
        proxy.storageinstance.add(new_master_storage)


def change_storage_cluster_entry(cluster_id: int, slave_ip: str, new_slave_ip: str):
    cluster = Cluster.objects.get(id=cluster_id)
    slave_storage = cluster.storageinstance_set.get(machine__ip=slave_ip)
    new_slave_storage = cluster.storageinstance_set.get(machine__ip=new_slave_ip)
    for be in slave_storage.bind_entry.all():
        be.storageinstance_set.remove(slave_storage)
        be.storageinstance_set.add(new_slave_storage)
    # 如果是standby节点，为了防止主节点故障dbHa切换后。从域名实际上指向的是主节点。需要从主节点读取域名并移除和添加
    if slave_storage.is_stand_by is True:
        master_storage = cluster.storageinstance_set.get(instance_inner_role=InstanceInnerRole.MASTER.value)
        for be in master_storage.bind_entry.filter(
            cluster_entry_type=ClusterEntryType.DNS.value, role=ClusterEntryRole.SLAVE_ENTRY.value
        ):
            be.storageinstance_set.remove(master_storage)
            be.storageinstance_set.add(new_slave_storage)
