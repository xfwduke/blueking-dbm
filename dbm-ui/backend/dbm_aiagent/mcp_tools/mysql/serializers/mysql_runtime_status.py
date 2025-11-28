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
from rest_framework import serializers


class MySQLInstanceAddressesSerializer(serializers.Serializer):
    addresses = serializers.ListField(child=serializers.CharField(help_text=_("MySQL 实例地址, 格式为 ip:port")),
                                      help_text=_("MySQL 实例地址列表, 每个地址的格式为 ip:port"),
                                      required=True)


class ShowProcesslistResponseSerializer(serializers.Serializer):
    address = serializers.CharField(help_text=_("MySQL 实例地址, 格式为 ip:port"))
    current_time = serializers.TimeField(help_text=_("MySQL 实例当前时间"))
