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

from backend.ticket import builders
from backend.ticket.builders.mysql.base import BaseMySQLSingleTicketFlowBuilder
from backend.ticket.builders.mysql.mysql_db_table_backup import (
    MySQLDBTableBackupDetailSerializer,
    MySQLDBTableBackupFlowParamBuilder,
)
from backend.ticket.constants import FlowRetryType, TicketType


@builders.BuilderFactory.register(TicketType.MYSQL_SINGLE_DB_TABLE_BACKUP)
class TenDBSingleDBTableBackupFlowBuilder(BaseMySQLSingleTicketFlowBuilder):
    serializer = MySQLDBTableBackupDetailSerializer
    inner_flow_builder = MySQLDBTableBackupFlowParamBuilder
    inner_flow_name = _("TenDBSingle 库表备份执行")
    retry_type = FlowRetryType.MANUAL_RETRY
