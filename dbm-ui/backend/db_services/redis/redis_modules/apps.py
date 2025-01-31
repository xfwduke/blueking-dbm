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
from django.db.models.signals import post_migrate

logger = logging.getLogger("root")


def init_default_modules(sender, **kwargs):
    from backend.db_services.redis.redis_modules.models import TbRedisModuleSupport

    try:
        TbRedisModuleSupport.init_default_modules()
    except Exception as err:  # pylint: disable=broad-except:
        logger.warning(f"init_default_modules occur error, {err}")


class InstanceConfig(AppConfig):
    default_auto_field = "django.db.models.BigAutoField"
    name = "backend.db_services.redis.redis_modules"

    def ready(self):
        post_migrate.connect(init_default_modules, sender=self)
