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
from collections import defaultdict

from django.utils.translation import ugettext as _
from pipeline.component_framework.component import Component

from backend.components.mysql_priv_manager.client import DBPrivManagerApi
from backend.db_meta.exceptions import ClusterNotExistException
from backend.db_meta.models import Cluster
from backend.flow.plugins.components.collections.common.base_service import BaseService
from backend.flow.utils.mysql.common.random_job_with_ticket_map import get_instance_with_random_job
from backend.flow.utils.mysql.get_mysql_sys_user import generate_mysql_tmp_user

logger = logging.getLogger("flow")


class AddTempUserForClusterService(BaseService):
    """
    为单据添加job的临时本地账号，操作目标实例
    单据是以集群维度来添加，如果单据涉及到集群，应该统一添加账号密码，以便后续操作方便
    """

    def _execute(self, data, parent_data, callback=None) -> bool:
        kwargs = data.get_one_of_inputs("kwargs")
        global_data = data.get_one_of_inputs("global_data")

        encrypt_switch_pwd = global_data["job_root_id"]

        bk_cloud_instance_map = defaultdict(list)
        for cluster_id in kwargs["cluster_ids"]:
            try:
                cluster = Cluster.objects.get(id=cluster_id, bk_biz_id=global_data["bk_biz_id"])
            except Cluster.DoesNotExist:
                raise ClusterNotExistException(
                    cluster_id=cluster_id, bk_biz_id=global_data["bk_biz_id"], message=_("集群不存在")
                )

            # 获取每套集群的所有需要添加临时的账号
            instance_list = get_instance_with_random_job(
                cluster=cluster, ticket_type=global_data.get("ticket_type", "test")
            )
            bk_cloud_instance_map[cluster.id].extend([ele["instance"] for ele in instance_list])

        err_cnt = 0
        for bk_cloud_id, instance_list in bk_cloud_instance_map.items():
            try:
                DBPrivManagerApi.add_priv_without_account_rule_v2(
                    {
                        "bk_cloud_id": bk_cloud_id,
                        "user": generate_mysql_tmp_user(global_data["job_root_id"]),
                        "psw": encrypt_switch_pwd,
                        "addresses": instance_list,
                    }
                )
                self.log_info(_("创建临时账号成功"))
            except Exception as e:  # pylint: disable=broad-except
                self.log_error(_("添加用户异常: {}").format(e))
                err_cnt += 1

        if err_cnt > 0:
            return False

        return True


class AddTempUserForClusterComponent(Component):
    name = __name__
    code = "add_job_temp_user"
    bound_service = AddTempUserForClusterService
