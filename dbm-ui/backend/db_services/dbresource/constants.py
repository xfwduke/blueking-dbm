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

from backend.db_meta.enums.spec import SpecClusterType, SpecMachineType
from backend.db_services.dbresource.handlers import (
    MongoDBShardSpecFilter,
    RedisClusterSpecFilter,
    TenDBClusterSpecFilter,
    TendisCacheSpecFilter,
    TendisPlusSpecFilter,
    TendisSSDSpecFilter,
)
from blue_krill.data_types.enum import EnumField, StructuredEnum

SWAGGER_TAG = _("资源池")

RESOURCE_IMPORT_TASK_FIELD = "{user}_resource_import_task_field"
RESOURCE_IMPORT_EXPIRE_TIME = 7 * 24 * 60 * 60


# 集群对应的规格筛选类
SPEC_FILTER_FACTORY = {
    SpecClusterType.TenDBCluster: {SpecMachineType.BACKEND: TenDBClusterSpecFilter},
    SpecClusterType.Redis: {
        SpecMachineType.TendisPredixyRedisCluster: RedisClusterSpecFilter,
        SpecMachineType.TendisPredixyTendisplusCluster: TendisPlusSpecFilter,
        SpecMachineType.TendisTwemproxyRedisInstance: TendisCacheSpecFilter,
        SpecMachineType.TwemproxyTendisSSDInstance: TendisSSDSpecFilter,
    },
    SpecClusterType.MongoDB: {SpecMachineType.MONGODB: MongoDBShardSpecFilter},
}


class ResourceOperation(str, StructuredEnum):
    import_hosts = EnumField("imported", _("导入主机"))
    consume_hosts = EnumField("consumed", _("消费主机"))


class ResourceGroupByEnum(str, StructuredEnum):
    DEVICE_CLASS = EnumField("device_class", _("按照机型聚合"))
    SPEC = EnumField("spec", _("按规格聚合"))
