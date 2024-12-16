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

from backend.components import DBPrivManagerApi
from backend.db_meta.exceptions import ClusterNotExistException
from backend.db_meta.models import Cluster
from backend.flow.plugins.components.collections.common.base_service import BaseService
from backend.flow.utils.mysql.common.random_job_with_ticket_map import get_instance_with_random_job
from backend.flow.utils.mysql.get_mysql_sys_user import generate_mysql_tmp_user

logger = logging.getLogger("flow")


class DropTempUserForClusterService(BaseService):
    """
    为单据删除job的临时本地账号，操作目标实例
    单据是以集群维度来删除
    """

    def _execute(self, data, parent_data, callback=None) -> bool:
        kwargs = data.get_one_of_inputs("kwargs")
        global_data = data.get_one_of_inputs("global_data")

        bk_cloud_instance_map = defaultdict(list)
        for cluster_id in kwargs["cluster_ids"]:
            # 获取每个cluster_id对应的对象
            try:
                cluster = Cluster.objects.get(id=cluster_id, bk_biz_id=global_data["bk_biz_id"])
            except Cluster.DoesNotExist:
                raise ClusterNotExistException(
                    cluster_id=cluster_id, bk_biz_id=global_data["bk_biz_id"], message=_("集群不存在")
                )

            instance_list = get_instance_with_random_job(
                cluster=cluster, ticket_type=global_data.get("ticket_type", "test")
            )
            bk_cloud_instance_map[cluster.bk_cloud_id].extend([ele["instance"] for ele in instance_list])

        err_cnt = 0
        for bk_cloud_id, instance_list in bk_cloud_instance_map.items():
            try:
                DBPrivManagerApi.drop_job_temp_account_v2(
                    {
                        "bk_cloud_id": bk_cloud_id,
                        "user": generate_mysql_tmp_user(global_data["job_root_id"]),
                        "addresses": instance_list,
                    }
                )
                self.log_info(_("删除临时账号成功"))
            except Exception as e:  # pylint: disable=broad-except
                self.log_error(_("删除用户接口异常: {}").format(e))
                err_cnt = err_cnt + 1

        if err_cnt > 0:
            return False

        return True


class DropTempUserForClusterComponent(Component):
    name = __name__
    code = "drop_job_temp_user"
    bound_service = DropTempUserForClusterService
