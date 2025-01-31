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

from django.apps import AppConfig
from django.db import IntegrityError
from django.db.models.signals import m2m_changed, post_migrate, post_save, pre_delete

logger = logging.getLogger("root")


def init_db_meta(sender, **kwargs):
    """初始化配置"""

    # 初始化城市配置
    from .models.city_map import BKCity, LogicalCity
    from .models.spec import Spec

    try:
        if not LogicalCity.objects.count():
            logical_city = LogicalCity.objects.create(name="default")
            BKCity.objects.create(bk_idc_city_name="default", logical_city=logical_city)
    except Exception as err:  # pylint: disable=broad-except:
        logger.warning(f"City init occur error: {err}, ignore...")

    # 初始化规格配置
    try:
        if not Spec.objects.count():
            Spec.init_spec()
    except (IntegrityError, Exception) as err:  # pylint: disable=broad-except:
        logger.warning(f"Spec init occur error: {err}, maybe already init, ignore...")


class DBMeta(AppConfig):
    default_auto_field = "django.db.models.BigAutoField"
    name = "backend.db_meta"

    def ready(self):
        from backend.db_meta.models import AppCache, ProxyInstance, StorageInstance
        from backend.db_meta.signals import update_cluster_status
        from backend.db_meta.utils import cache_appcache_data

        post_migrate.connect(cache_appcache_data, sender=self)
        post_save.connect(cache_appcache_data, sender=AppCache)

        post_migrate.connect(init_db_meta, sender=self)
        # 当实例进行修改或者删除时，更新集群状态
        post_save.connect(update_cluster_status, sender=StorageInstance)
        post_save.connect(update_cluster_status, sender=ProxyInstance)
        pre_delete.connect(update_cluster_status, sender=StorageInstance)
        pre_delete.connect(update_cluster_status, sender=ProxyInstance)
        m2m_changed.connect(update_cluster_status, sender=StorageInstance.cluster.through)
        m2m_changed.connect(update_cluster_status, sender=ProxyInstance.cluster.through)
