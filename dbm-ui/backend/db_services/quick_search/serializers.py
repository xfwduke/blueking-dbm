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

from django.utils.translation import ugettext as _
from rest_framework import serializers

from backend.configuration.constants import DBType
from backend.db_services.quick_search.constants import FilterType, ResourceType


class QuickSearchSerializer(serializers.Serializer):
    bk_biz_ids = serializers.ListSerializer(help_text=_("业务列表"), child=serializers.IntegerField(), allow_empty=True)
    db_types = serializers.ListSerializer(
        help_text=_("数据库类型列表"),
        child=serializers.ChoiceField(help_text=_("数据库类型"), choices=DBType.get_choices(), required=False),
        allow_empty=True,
    )
    resource_types = serializers.ListSerializer(
        help_text=_("资源列表"),
        child=serializers.ChoiceField(help_text=_("资源类型"), choices=ResourceType.get_choices(), required=False),
        allow_empty=True,
    )
    filter_type = serializers.ChoiceField(help_text=_("检索类型"), choices=FilterType.get_choices(), required=False)
    keyword = serializers.CharField(help_text=_("关键字过滤"), required=False)
    short_code = serializers.CharField(help_text=_("短码"), required=False)
    limit = serializers.IntegerField(help_text=_("分页大小"), required=False, default=10)

    def validate(self, attrs):
        # keyword 和 short_code 不能同时为空
        if not attrs.get("keyword") and not attrs.get("short_code"):
            raise serializers.ValidationError(_("keyword 和 short_code 不能同时为空"))
        return attrs
