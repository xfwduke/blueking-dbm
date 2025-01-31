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

import uuid

from backend.configuration.constants import DBType
from backend.db_meta.enums.cluster_type import ClusterType
from backend.flow.models import StateType
from backend.tests.mock_data import constant
from backend.ticket.constants import TicketType

ROOT_ID = "a884b5422f7111ed830cc2afcf9e926b"
SQL_IMPORT_NODE_ID = "a651615616516dwqd156dq6616516qd"
SQL_IMPORT_VERSION_ID = "d516156156qwd161651665161656"
BK_USERNAME = "admin"
SN = "NO2019090519542603"
BK_BIZ_ID = constant.BK_BIZ_ID

MYSQL_AUTHORIZE_TICKET_DATA = {
    "bk_biz_id": constant.BK_BIZ_ID,
    "details": {
        "authorize_uid": uuid.uuid1().hex,
        "authorize_data": {
            "user": "admin",
            "access_dbs": ["dbnew", "user"],
            "source_ips": [{"ip": "1.1.1.1", "bk_host_id": 1}, {"ip": "2.2.2.2", "bk_host_id": 2}],
            "target_instances": ["gamedb.privtest55.blueking.db"],
            "cluster_type": "tendbha",
        },
    },
    "remark": "",
    "ticket_type": "MYSQL_AUTHORIZE_RULES",
}

MYSQL_FULL_BACKUP_TICKET_DATA = {
    "bk_biz_id": constant.BK_BIZ_ID,
    "details": {
        "backup_type": "logical",
        "file_tag": "DBFILE1M",
        "infos": [{"cluster_id": 1, "backup_local": "master"}],
    },
    "remark": "",
    "ticket_type": "MYSQL_HA_FULL_BACKUP",
}

MYSQL_PERMISSION_ACCOUNT = {
    "items": [
        {
            "account": {
                "bk_biz_id": 3,
                "user": "admin",
                "creator": "admin",
                "create_time": "2023-08-07T21:34:30Z",
                "account_id": 1,
            },
            "rules": [
                {
                    "account_id": 1,
                    "bk_biz_id": 3,
                    "creator": "",
                    "create_time": "2023-08-10T07:09:36Z",
                    "rule_id": 4,
                    "dbname": "ddddd",
                    "priv": "select,insert,update,delete",
                }
            ],
        }
    ]
}
MYSQL_ITSM_AUTHORIZE_TICKET_DATA = [
    {
        "user": "admin",
        "index": 0,
        "message": "",
        "operator": "admin",
        "bk_biz_id": 3,
        "source_ips": ["127.0.0.1"],
        "cluster_type": "tendbha",
        "account_rules": [{"dbname": "ddddd", "bk_biz_id": 3}],
    }
]
MYSQL_AUTHORIZE_CLONE_CLIENT_TICKET_DATA = {
    "bk_biz_id": constant.BK_BIZ_ID,
    "details": {
        "clone_uid": "80fc459ae1d51xxxx17626xxxb38e5",
        "clone_data_list": [
            {"module": "Test/Server/", "source": "127.0.0.1", "target": ["127.0.0.2"], "bk_cloud_id": 0}
        ],
        "clone_type": "client",
    },
    "remark": "",
    "ticket_type": "MYSQL_CLIENT_CLONE_RULES",
}

MYSQL_CLONE_CLIENT_TICKET_CONFIG = {
    "bk_biz_id": constant.BK_BIZ_ID,
    "ticket_type": TicketType.MYSQL_CLIENT_CLONE_RULES,
    "configs": {"need_itsm": True, "need_manual_confirm": True},
    "editable": 1,
    "group": DBType.MySQL,
}

MYSQL_SINGLE_APPLY_TICKET_DATA = {
    "bk_biz_id": BK_BIZ_ID,
    "remark": "",
    "ticket_type": "MYSQL_SINGLE_APPLY",
    "details": {
        "bk_cloud_id": 0,
        "city_code": "南京",
        "db_app_abbr": "blueking",
        "spec": "SA2.SMALL4",
        "db_module_id": 1,
        "cluster_count": 1,
        "charset": "",
        "mysql_port": 20000,
        "proxy_port": 10000,
        "domains": [{"key": "kio"}],
        "disaster_tolerance_level": "same_city_cross_zone",
    },
}

MYSQL_TENDBHA_TICKET_DATA = {
    "bk_biz_id": BK_BIZ_ID,
    "ticket_type": "MYSQL_HA_APPLY",
    "remark": "",
    "details": {
        "bk_cloud_id": 0,
        "city_code": "南京",
        "db_app_abbr": "blueking",
        "spec": "SA2.SMALL4",
        "db_module_id": 1,
        "cluster_count": 1,
        "charset": "",
        "mysql_port": 20000,
        "proxy_port": 10000,
        "domains": [{"key": "kio"}],
        "disaster_tolerance_level": "SAME_SUBZONE_CROSS_SWTICH",
        "resource_spec": {
            "proxy": {
                "affinity": "CROS_SUBZONE",
                "location_spec": {"city": "default", "sub_zone_ids": []},
                "spec_name": "spec_redis",
                "spec_id": 1,
                "count": 1,
                "spec_cluster_type": "TwemproxyRedisInstance",
                "spec_machine_type": "tendisssd",
            }
        },
    },
}

APPLY_RESOURCE_RETURN_DATA = {
    "proxy": [
        {
            "ip": "127.0.0.1",
            "bk_host_id": 2000058793,
            "bk_cloud_id": 0,
            "bk_biz_id": BK_BIZ_ID,
            "bk_cpu": 2000058793,
            "bk_mem": 2000058793,
        },
        {
            "ip": "127.0.0.1",
            "bk_host_id": 2000058794,
            "bk_cloud_id": 0,
            "bk_biz_id": 2000058793,
            "bk_cpu": 2000058793,
            "bk_mem": 2000058793,
        },
    ],
    "backend": [
        {
            "master": {
                "ip": "127.0.0.1",
                "bk_host_id": 2000058795,
                "bk_biz_id": BK_BIZ_ID,
                "bk_cpu": 2000058793,
                "bk_mem": 2000058793,
                "bk_cloud_id": 0,
            },
            "slave": {
                "ip": "127.0.0.1",
                "bk_host_id": 2000058796,
                "bk_biz_id": BK_BIZ_ID,
                "bk_cpu": 2000058793,
                "bk_mem": 2000058793,
                "bk_cloud_id": 0,
            },
        },
    ],
}


SQL_IMPORT_DATA = {
    "bk_biz_id": BK_BIZ_ID,
    "details": {
        "root_id": ROOT_ID,
        "charset": "default",
        "force": False,
        "path": "/bk-dbm/mysql/sqlfile",
        "cluster_ids": [110],
        "execute_objects": [
            {
                "sql_files": ["bar.sql", "foo.sql"],
                "dbnames": ["db_log%"],
                "ignore_dbnames": ["db1", "db2"],
                "import_mode": "file",
            }
        ],
        "ticket_mode": {"mode": "auto"},
        "backup": [],
        "highrisk_warnings": "",
        "bk_biz_id": BK_BIZ_ID,
        "created_by": "admin",
    },
    "remark": "",
    "ticket_type": "MYSQL_IMPORT_SQLFILE",
}
SQL_IMPORT_TICKET_DATA = {
    "bk_biz_id": constant.BK_BIZ_ID,
    "details": {"root_id": ROOT_ID},
    "remark": "",
    "ticket_type": "MYSQL_IMPORT_SQLFILE",
}

DB_MODULE_DATA = {
    "creator": "admin",
    "create_at": "2022-07-28 07:09:46",
    "updater": "admin",
    "update_at": "2022-07-29 07:09:46",
    "bk_biz_id": constant.BK_BIZ_ID,
    "db_module_name": "blueking-module",
    "db_module_id": 1,
    "cluster_type": ClusterType.TenDBSingle,
    "alias_name": "",
}

FLOW_TREE_DATA = {
    "bk_biz_id": constant.BK_BIZ_ID,
    "uid": "REQ20200831000005",
    "ticket_type": TicketType.MYSQL_SINGLE_APPLY.name,
    "root_id": ROOT_ID,
    "tree": {"activities": {"SQL_IMPORT_NODE_ID": {"component": {"code": 1}}}},
    "status": StateType.FINISHED.value,
    "created_by": BK_USERNAME,
    "created_at": "2022-07-28 07:09:46",
    "updated_at": "2022-07-29 07:09:46",
}

SQL_IMPORT_FLOW_NODE_DATA = {
    "uid": 1,
    "root_id": ROOT_ID,
    "node_id": SQL_IMPORT_NODE_ID,
    "version_id": SQL_IMPORT_VERSION_ID,
    "status": StateType.FINISHED.value,
}
