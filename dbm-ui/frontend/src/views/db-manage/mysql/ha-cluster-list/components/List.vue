<!--
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-DB管理系统(BlueKing-BK-DBM) available.
 *
 * Copyright (C) 2017-2023 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License athttps://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed
 * on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for
 * the specific language governing permissions and limitations under the License.
-->

<template>
  <div class="mysql-ha-cluster-list-page">
    <div class="operation-box">
      <AuthButton
        v-db-console="'mysql.haClusterList.instanceApply'"
        action-id="mysql_apply"
        theme="primary"
        @click="handleApply">
        {{ t('申请实例') }}
      </AuthButton>
      <ClusterBatchOperation
        v-db-console="'mysql.haClusterList.batchOperation'"
        class="ml-8"
        :cluster-type="ClusterTypes.TENDBHA"
        :selected="selected"
        @success="handleBatchOperationSuccess" />
      <BkButton
        v-db-console="'mysql.haClusterList.importAuthorize'"
        class="ml-8"
        @click="handleShowExcelAuthorize">
        {{ t('导入授权') }}
      </BkButton>
      <DropdownExportExcel
        v-db-console="'mysql.haClusterList.export'"
        :ids="selectedIds"
        type="tendbha" />
      <ClusterIpCopy
        v-db-console="'mysql.haClusterList.batchCopy'"
        :selected="selected" />
      <DbSearchSelect
        :data="searchSelectData"
        :get-menu-list="getMenuList"
        :model-value="searchValue"
        :placeholder="t('请输入或选择条件搜索')"
        unique-select
        :validate-values="validateSearchValues"
        @change="handleSearchValueChange" />
    </div>
    <div
      class="table-wrapper"
      :class="{ 'is-shrink-table': isStretchLayoutOpen }">
      <DbTable
        ref="tableRef"
        :columns="columns"
        :data-source="getTendbhaList"
        :line-height="80"
        releate-url-query
        :row-class="setRowClass"
        selectable
        :settings="settings"
        :show-overflow="false"
        @clear-search="clearSearchValue"
        @column-filter="columnFilterChange"
        @column-sort="columnSortChange"
        @selection="handleSelection"
        @setting-change="updateTableSettings" />
    </div>
  </div>
  <!-- 集群授权 -->
  <ClusterAuthorize
    v-model="authorizeState.isShow"
    :account-type="AccountTypes.MYSQL"
    :cluster-types="[ClusterTypes.TENDBHA, 'tendbhaSlave']"
    :selected="authorizeState.selected"
    @success="handleClearSelected" />
  <!-- excel 导入授权 -->
  <ExcelAuthorize
    v-model:is-show="isShowExcelAuthorize"
    :cluster-type="ClusterTypes.TENDBHA" />
  <CreateSubscribeRuleSlider
    v-model="showCreateSubscribeRuleSlider"
    :selected-clusters="selectedClusterList"
    show-tab-panel />
  <ClusterExportData
    v-if="currentData"
    v-model:is-show="showDataExportSlider"
    :data="currentData"
    :ticket-type="TicketTypes.MYSQL_DUMP_DATA" />
</template>

<script setup lang="tsx">
  import { Message } from 'bkui-vue';
  import type { ISearchItem } from 'bkui-vue/lib/search-select/utils';
  import { useI18n } from 'vue-i18n';

  import type { MySQLFunctions } from '@services/model/function-controller/functionController';
  import TendbhaModel from '@services/model/mysql/tendbha';
  import {
    getTendbhaInstanceList,
    getTendbhaList,
  } from '@services/source/tendbha';
  import { getUserList } from '@services/source/user';

  import {
    useCopy,
    useLinkQueryColumnSerach,
    useStretchLayout,
    useTableSettings,
  } from '@hooks';

  import {
    useFunController,
    useGlobalBizs,
  } from '@stores';

  import {
    AccountTypes,
    ClusterTypes,
    DBTypes,
    TicketTypes,
    UserPersonalSettings,
  } from '@common/const';

  import DbStatus from '@components/db-status/index.vue';
  import DbTable from '@components/db-table/index.vue';
  import MoreActionExtend from '@components/more-action-extend/Index.vue';
  import TextOverflowLayout from '@components/text-overflow-layout/Index.vue';

  import ClusterAuthorize from '@views/db-manage/common/cluster-authorize/Index.vue';
  import ClusterBatchOperation from '@views/db-manage/common/cluster-batch-opration/Index.vue'
  import ClusterCapacityUsageRate from '@views/db-manage/common/cluster-capacity-usage-rate/Index.vue'
  import EditEntryConfig, { type ClusterEntryInfo } from '@views/db-manage/common/cluster-entry-config/Index.vue';
  import ClusterExportData from '@views/db-manage/common/cluster-export-data/Index.vue'
  import ClusterIpCopy from '@views/db-manage/common/cluster-ip-copy/Index.vue';
  import DropdownExportExcel from '@views/db-manage/common/dropdown-export-excel/index.vue';
  import ExcelAuthorize from '@views/db-manage/common/ExcelAuthorize.vue';
  import { useOperateClusterBasic } from '@views/db-manage/common/hooks';
  import OperationBtnStatusTips from '@views/db-manage/common/OperationBtnStatusTips.vue';
  import RenderCellCopy from '@views/db-manage/common/render-cell-copy/Index.vue';
  import RenderHeadCopy from '@views/db-manage/common/render-head-copy/Index.vue';
  import RenderInstances from '@views/db-manage/common/render-instances/RenderInstances.vue';
  import RenderOperationTag from '@views/db-manage/common/RenderOperationTagNew.vue';
  import CreateSubscribeRuleSlider from '@views/db-manage/mysql/dumper/components/create-rule/Index.vue';

  import {
    getMenuListSearch,
    getSearchSelectorParams,
    isRecentDays,
  } from '@utils';

  import RenderEntries from './RenderEntries.vue';

  interface ColumnData {
    cell: string,
    data: TendbhaModel
  }

  const clusterId = defineModel<number>('clusterId');

  // 设置行样式
  const setRowClass = (row: TendbhaModel) => {
    const classList = [row.isOffline ? 'is-offline' : ''];
    const newClass = isRecentDays(row.create_at, 24 * 3) ? 'is-new-row' : '';
    classList.push(newClass);
    if (row.id === clusterId.value) {
      classList.push('is-selected-row');
    }
    return classList.filter(cls => cls).join(' ');
  };

  const route = useRoute();
  const router = useRouter();
  const globalBizsStore = useGlobalBizs();
  const funControllerStore = useFunController();
  const copy = useCopy();
  const { t, locale } = useI18n();
  const { handleDisableCluster, handleEnableCluster, handleDeleteCluster } = useOperateClusterBasic(
    ClusterTypes.TENDBHA,
    {
      onSuccess: () => fetchData(),
    },
  );
  const {
    isOpen: isStretchLayoutOpen,
    splitScreen: stretchLayoutSplitScreen,
  } = useStretchLayout();

  const {
    columnAttrs,
    searchAttrs,
    searchValue,
    sortValue,
    columnCheckedMap,
    batchSearchIpInatanceList,
    columnFilterChange,
    columnSortChange,
    clearSearchValue,
    validateSearchValues,
    handleSearchValueChange,
  } = useLinkQueryColumnSerach({
    searchType: ClusterTypes.TENDBHA,
    attrs: [
      'bk_cloud_id',
      'db_module_id',
      'major_version',
      'region',
      'time_zone',
    ],
    fetchDataFn: () => fetchData(),
    defaultSearchItem: {
      name: t('访问入口'),
      id: 'domain',
    }
  });

  const tableRef = ref<InstanceType<typeof DbTable>>();
  const isShowExcelAuthorize = ref(false);
  const isInit = ref(false);
  const showCreateSubscribeRuleSlider = ref(false);
  const showDataExportSlider = ref(false)
  const selectedClusterList = ref<ColumnData['data'][]>([]);
  const currentData = ref<ColumnData['data']>();

  const selected = ref<TendbhaModel[]>([])
  /** 集群授权 */
  const authorizeState = reactive({
    isShow: false,
    selected: [] as TendbhaModel[],
  });

  const isCN = computed(() => locale.value === 'zh-cn');
  const hasSelected = computed(() => selected.value.length > 0);
  const selectedIds = computed(() => selected.value.map(item => item.id));

  const searchSelectData = computed(() => [
    {
      name: t('访问入口'),
      id: 'domain',
      multiple: true,
    },
    {
      name: t('IP 或 IP:Port'),
      id: 'instance',
      multiple: true,
    },
    {
      name: 'ID',
      id: 'id',
    },
    {
      name: t('集群名称'),
      id: 'name',
    },
    {
      name: t('管控区域'),
      id: 'bk_cloud_id',
      multiple: true,
      children: searchAttrs.value.bk_cloud_id,
    },
    {
      name: t('状态'),
      id: 'status',
      multiple: true,
      children: [
        {
          id: 'normal',
          name: t('正常'),
        },
        {
          id: 'abnormal',
          name: t('异常'),
        },
      ],
    },
    {
      name: t('所属DB模块'),
      id: 'db_module_id',
      multiple: true,
      children: searchAttrs.value.db_module_id,
    },
    {
      name: t('版本'),
      id: 'major_version',
      multiple: true,
      children: searchAttrs.value.major_version,
    },
    {
      name: t('地域'),
      id: 'region',
      multiple: true,
      children: searchAttrs.value.region,
    },
    {
      name: t('创建人'),
      id: 'creator',
    },
    {
      name: t('时区'),
      id: 'time_zone',
      multiple: true,
      children: searchAttrs.value.time_zone,
    },
  ]);

  const tableOperationWidth = computed(() => {
    if (!isStretchLayoutOpen.value) {
      return isCN.value ? 270 : 280;
    }
    return 60;
  });

  const isShowDumperEntry = computed(() => {
    const currentKey = `dumper_biz_${globalBizsStore.currentBizId}` as MySQLFunctions;
    return funControllerStore.funControllerData.mysql.children[currentKey];
  });

  const entrySort = (data: ClusterEntryInfo[]) => data.sort(a => a.role === 'master_entry' ? -1 : 1);

  const columns = computed(() => [
    {
      label: 'ID',
      field: 'id',
      fixed: 'left',
      width: 100,
    },
    {
      label: t('主访问入口'),
      field: 'master_domain',
      fixed: 'left',
      minWidth: 280,
      showOverflowTooltip: false,
      renderHead: () => (
        <RenderHeadCopy
          hasSelected={hasSelected.value}
          onHandleCopySelected={handleCopySelected}
          onHandleCopyAll={handleCopyAll}
          config={
            [
              {
                field: 'master_domain',
                label: t('域名')
              },
              {
                field: 'masterDomainDisplayName',
                label: t('域名:端口')
              }
            ]
          }
        >
          {t('主访问入口')}
        </RenderHeadCopy>
      ),
      render: ({ data }: ColumnData) => (
        <TextOverflowLayout>
          {{
            default: () => (
              <auth-button
                action-id="mysql_view"
                resource={data.id}
                permission={data.permission.mysql_view}
                text
                theme="primary"
                onClick={() => handleToDetails(data.id)}>
                {data.masterDomainDisplayName}
              </auth-button>
            ),
            append: () => (
              <>
                {
                  data.operationTagTips.map(item => <RenderOperationTag class="cluster-tag ml-4" data={item}/>)
                }
                {
                  data.isOffline && !data.isStarting && (
                    <bk-tag
                      class="ml-4"
                      size="small">
                      {t('已禁用')}
                    </bk-tag>
                  )
                }
                {
                  data.isNew && (
                    <bk-tag
                      theme="success"
                      size="small"
                      class="ml-4">
                      NEW
                    </bk-tag>
                  )
                }
                <RenderCellCopy copyItems={
                  [
                    {
                      value: data.master_domain,
                      label: t('域名')
                    },
                    {
                      value: data.masterDomainDisplayName,
                      label: t('域名:端口')
                    }
                  ]
                } />
                <span v-db-console="mysql.haClusterList.modifyEntryConfiguration">
                  <EditEntryConfig
                    id={data.id}
                    bizId={data.bk_biz_id}
                    permission={data.permission.access_entry_edit}
                    resource={DBTypes.MYSQL}
                    sort={entrySort}
                    onSuccess={fetchData}>
                      {{
                        prepend: ({ data: cluster }: { data: ClusterEntryInfo } ) =>
                          cluster.role === 'master_entry' ?
                            <bk-tag size="small" theme="success">{ t('主') }</bk-tag>
                            : <bk-tag size="small" theme="info">{ t('从') }</bk-tag>,
                      }}
                  </EditEntryConfig>
                </span>
              </>
            ),
          }}
        </TextOverflowLayout>
      ),
    },
    {
      label: t('集群名称'),
      field: 'cluster_name',
      minWidth: 200,
      showOverflowTooltip: false,
      renderHead: () => (
        <RenderHeadCopy
          hasSelected={hasSelected.value}
          onHandleCopySelected={handleCopySelected}
          onHandleCopyAll={handleCopyAll}
          config={
            [
              {
                field: 'cluster_name'
              },
            ]
          }
        >
          {t('集群名称')}
        </RenderHeadCopy>
      ),
      render: ({ data }: ColumnData) => (
        <TextOverflowLayout>
          {{
            default: () => data.cluster_name,
            append: () => (
              <>
                <span v-db-console="mysql.haClusterList.modifyEntryConfiguration">
                  <EditEntryConfig
                    id={data.id}
                    bizId={data.bk_biz_id}
                    permission={data.permission.access_entry_edit}
                    resource={DBTypes.MYSQL}
                    sort={entrySort}
                    onSuccess={fetchData}>
                      {{
                        prepend: ({ data: cluster }: { data: ClusterEntryInfo } ) =>
                          cluster.role === 'master_entry' ?
                            <bk-tag size="small" theme="success">{ t('主') }</bk-tag>
                            : <bk-tag size="small" theme="info">{ t('从') }</bk-tag>,
                      }}
                  </EditEntryConfig>
                </span>
              </>
            ),
          }}
        </TextOverflowLayout>
      ),
    },
    {
      label: t('状态'),
      field: 'status',
      width: 90,
      filter: {
        list: [
          {
            value: 'normal',
            text: t('正常'),
          },
          {
            value: 'abnormal',
            text: t('异常'),
          },
        ],
        checked: columnCheckedMap.value.status,
      },
      render: ({ data }: ColumnData) => {
        const info = data.status === 'normal' ? { theme: 'success', text: t('正常') } : { theme: 'danger', text: t('异常') };
        return <DbStatus theme={info.theme}>{info.text}</DbStatus>;
      },
    },
    {
      label: t('容量使用率'),
      field: 'cluster_stats',
      width: 240,
      showOverflowTooltip: false,
      render: ({ data }: ColumnData) => <ClusterCapacityUsageRate clusterStats={data.cluster_stats} />
    },
    {
      label: t('从访问入口'),
      field: 'slave_domain',
      minWidth: 280,
      width: 280,
      showOverflowTooltip: false,
      renderHead: () => (
        <RenderHeadCopy
          hasSelected={hasSelected.value}
          onHandleCopySelected={handleCopySelected}
          onHandleCopyAll={handleCopyAll}
          config={
            [
              {
                field: 'slave_domain',
                label: t('域名')
              },
              {
                field: 'slaveDomainDisplayName',
                label: t('域名:端口')
              }
            ]
          }
        >
          {t('从访问入口')}
        </RenderHeadCopy>
      ),
      render: ({ data }: ColumnData) => <RenderEntries data={data.slaveEntryDisplayList}>
        {{
          append: ({ index }: { index: number }) => index === 0 && (
            <>
              <RenderCellCopy copyItems={
                [
                  {
                    value: data.slaveEntryList.join('\n'),
                    label: t('域名')
                  },
                  {
                    value: data.slaveEntryDisplayList.join('\n'),
                    label: t('域名:端口')
                  }
                ]
              } />
              <span v-db-console="mysql.haClusterList.modifyEntryConfiguration">
                <EditEntryConfig
                  id={data.id}
                  bizId={data.bk_biz_id}
                  permission={data.permission.access_entry_edit}
                  resource={DBTypes.MYSQL}
                  sort={entrySort}
                  onSuccess={fetchData}>
                    {{
                      prepend: ({ data: cluster }: { data: ClusterEntryInfo } ) =>
                        cluster.role === 'master_entry' ?
                          <bk-tag size="small" theme="success">{ t('主') }</bk-tag>
                          : <bk-tag size="small" theme="info">{ t('从') }</bk-tag>,
                    }}
                </EditEntryConfig>
              </span>
            </>)
        }}
        </RenderEntries>
    },
    {
      label: 'Proxy',
      field: 'proxies',
      width: 200,
      minWidth: 200,
      showOverflowTooltip: false,
      renderHead: () => (
        <RenderHeadCopy
          hasSelected={hasSelected.value}
          onHandleCopySelected={(field) => handleCopySelected(field, 'proxies')}
          onHandleCopyAll={(field) => handleCopyAll(field, 'proxies')}
          config={
            [
              {
                label: 'IP',
                field: 'ip'
              },
              {
                label: t('实例'),
                field: 'instance'
              }
            ]
          }
        >
          {'Proxy'}
        </RenderHeadCopy>
      ),
      render: ({ data }: ColumnData) => (
        <RenderInstances
          highlightIps={batchSearchIpInatanceList.value}
          data={data.proxies || []}
          title={t('【inst】实例预览', { inst: data.master_domain, title: 'Proxy' })}
          role="proxy"
          clusterId={data.id}
          dataSource={getTendbhaInstanceList}
        />
      ),
    },
    {
      label: 'Master',
      field: 'masters',
      width: 200,
      minWidth: 200,
      showOverflowTooltip: false,
      renderHead: () => (
        <RenderHeadCopy
          hasSelected={hasSelected.value}
          onHandleCopySelected={(field) => handleCopySelected(field, 'masters')}
          onHandleCopyAll={(field) => handleCopyAll(field, 'masters')}
          config={
            [
              {
                label: 'IP',
                field: 'ip'
              },
              {
                label: t('实例'),
                field: 'instance'
              }
            ]
          }
        >
          {'Master'}
        </RenderHeadCopy>
      ),
      render: ({ data }: ColumnData) => (
        <RenderInstances
          highlightIps={batchSearchIpInatanceList.value}
          data={data.masters}
          title={t('【inst】实例预览', { inst: data.master_domain, title: 'Master' })}
          role="proxy"
          clusterId={data.id}
          dataSource={getTendbhaInstanceList}
        />
      ),
    },
    {
      label: 'Slave',
      field: 'slaves',
      width: 200,
      minWidth: 200,
      showOverflowTooltip: false,
      renderHead: () => (
        <RenderHeadCopy
          hasSelected={hasSelected.value}
          onHandleCopySelected={(field) => handleCopySelected(field, 'slaves')}
          onHandleCopyAll={(field) => handleCopyAll(field, 'slaves')}
          config={
            [
              {
                label: 'IP',
                field: 'ip'
              },
              {
                label: t('实例'),
                field: 'instance'
              }
            ]
          }
        >
          {'Slave'}
        </RenderHeadCopy>
      ),
      render: ({ data }: ColumnData) => (
        <RenderInstances
          highlightIps={batchSearchIpInatanceList.value}
          data={(data.slaves || []).sort((a, b) => Number(b.is_stand_by) - Number(a.is_stand_by))}
          title={t('【inst】实例预览', { inst: data.master_domain, title: 'Slave' })}
          role="slave"
          clusterId={data.id}
          dataSource={getTendbhaInstanceList}
        >
          {{
            append: ({ data: instance }: { data: TendbhaModel['slaves'][number] }) =>
              data.slaves.length > 1 && instance.is_stand_by && (<bk-tag class="is-stand-by" size="small">Standby</bk-tag>)
          }}
        </RenderInstances>
      ),
    },
    {
      label: t('所属DB模块'),
      field: 'db_module_id',
      width: 140,
      filter: {
        list: columnAttrs.value.db_module_id,
        checked: columnCheckedMap.value.db_module_id,
      },
      render: ({ data }: ColumnData) => <span>{data.db_module_name || '--'}</span>,
    },
    {
      label: t('版本'),
      field: 'major_version',
      minWidth: 100,
      filter: {
        list: columnAttrs.value.major_version,
        checked: columnCheckedMap.value.major_version,
      },
      render: ({ cell }: ColumnData) => <span>{cell || '--'}</span>,
    },
    {
        label: t('容灾要求'),
        field: 'disaster_tolerance_level',
        minWidth: 100,
        render: ({ data }: ColumnData) => data.disasterToleranceLevelName || '--',
    },
    {
      label: t('地域'),
      field: 'region',
      minWidth: 100,
      filter: {
        list: columnAttrs.value.region,
        checked: columnCheckedMap.value.region,
      },
      render: ({ cell }: ColumnData) => <span>{cell || '--'}</span>,
    },
    {
      label: t('管控区域'),
      field: 'bk_cloud_id',
      filter: {
        list: columnAttrs.value.bk_cloud_id,
        checked: columnCheckedMap.value.bk_cloud_id,
      },
      width: 90,
      render: ({ data }: ColumnData) =>  data.bk_cloud_name ? `${data.bk_cloud_name}[${data.bk_cloud_id}]` : '--',
    },
    {
      label: t('创建人'),
      field: 'creator',
      width: 140,
      render: ({ cell }: ColumnData) => <span>{cell || '--'}</span>,
    },
    {
      label: t('部署时间'),
      field: 'create_at',
      width: 200,
      sort: true,
      render: ({ data }: ColumnData) => <span>{data.createAtDisplay || '--'}</span>,
    },
    {
      label: t('时区'),
      field: 'cluster_time_zone',
      width: 100,
      filter: {
        list: columnAttrs.value.time_zone,
        checked: columnCheckedMap.value.time_zone,
      },
      render: ({ cell }: ColumnData) => <span>{cell || '--'}</span>,
    },
    {
      label: t('操作'),
      field: '',
      width: tableOperationWidth.value,
      fixed: isStretchLayoutOpen.value ? false : 'right',
      showOverflowTooltip: false,
      render: ({ data }: ColumnData) => (
        <>
          <bk-button
            v-db-console="mysql.haClusterList.authorize"
            text
            theme="primary"
            class="mr-8"
            disabled={data.isOffline}
            onClick={() => handleShowAuthorize([data])}>
            { t('授权') }
          </bk-button>
          <auth-button
            v-db-console="mysql.haClusterList.webconsole"
            action-id="mysql_webconsole"
            resource={data.id}
            permission={data.permission.mysql_webconsole}
            disabled={data.isOffline}
            text
            theme="primary"
            class="mr-8"
            onClick={() => handleGoWebconsole(data.id)}>
            Webconsole
          </auth-button>
          <auth-button
            v-db-console="mysql.haClusterList.exportData"
            action-id="mysql_dump_data"
            resource={data.id}
            permission={data.permission.mysql_dump_data}
            disabled={data.isOffline}
            text
            theme="primary"
            class="mr-16"
            onClick={() => handleShowDataExportSlider(data)}>
            { t('导出数据') }
          </auth-button>
          <MoreActionExtend v-db-console="mysql.haClusterList.moreOperation">
            {{
              default: () => <>
                {isShowDumperEntry.value && (
                  <bk-dropdown-item v-db-console="mysql.dataSubscription">
                    <auth-button
                      action-id="tbinlogdumper_install"
                      resource={data.id}
                      disabled={data.isOffline}
                      permission={data.permission.tbinlogdumper_install}
                      text
                      class="mr-8"
                      onClick={() => handleShowCreateSubscribeRuleSlider(data)}>
                      { t('数据订阅') }
                    </auth-button>
                  </bk-dropdown-item>
                )}
                {data.isOnline ? (
                  <bk-dropdown-item v-db-console="mysql.haClusterList.disable">
                    <OperationBtnStatusTips data={data}>
                      <auth-button
                        text
                        disabled={data.operationDisabled}
                        class="mr-8"
                        action-id="mysql_enable_disable"
                        permission={data.permission.mysql_enable_disable}
                        resource={data.id}
                        onClick={() => handleDisableCluster([data])}>
                        { t('禁用') }
                      </auth-button>
                    </OperationBtnStatusTips>
                  </bk-dropdown-item>
                ) : (
                  <bk-dropdown-item v-db-console="mysql.haClusterList.enable">
                    <OperationBtnStatusTips data={data}>
                      <auth-button
                        text
                        disabled={data.isStarting}
                        class="mr-8"
                        action-id="mysql_enable_disable"
                        permission={data.permission.mysql_enable_disable}
                        resource={data.id}
                        onClick={() => handleEnableCluster([data])}>
                        { t('启用') }
                      </auth-button>
                    </OperationBtnStatusTips>
                  </bk-dropdown-item>
                )}
                <bk-dropdown-item v-db-console="mysql.haClusterList.delete">
                  <OperationBtnStatusTips data={data}>
                    <auth-button
                      v-bk-tooltips={{
                        disabled: data.isOffline,
                        content: t('请先禁用集群')
                      }}
                      text
                      disabled={data.isOnline || Boolean(data.operationTicketId)}
                      class="mr-8"
                      action-id="mysql_destroy"
                      permission={data.permission.mysql_destroy}
                      resource={data.id}
                      onClick={() => handleDeleteCluster([data])}>
                      { t('删除') }
                    </auth-button>
                  </OperationBtnStatusTips>
                </bk-dropdown-item>
              </>
            }}
          </MoreActionExtend>
        </>
      ),
    },
  ]);

  const defaultSettings = {
    fields: (columns.value || []).filter(item => item.field).map(item => ({
      label: item.label as string,
      field: item.field as string,
      disabled: ['master_domain'].includes(item.field as string),
    })),
    checked: [
      'master_domain',
      'status',
      'cluster_stats',
      'slave_domain',
      'proxies',
      'masters',
      'slaves',
      'db_module_id',
      'major_version',
      'disaster_tolerance_level',
      'region',
      'bk_cloud_id'
    ],
    showLineHeight: false,
    trigger: 'manual' as const,
  };

  const {
    settings,
    updateTableSettings,
  } = useTableSettings(UserPersonalSettings.TENDBHA_TABLE_SETTINGS, defaultSettings);

  const getMenuList = async (item: ISearchItem | undefined, keyword: string) => {
    if (item?.id !== 'creator' && keyword) {
      return getMenuListSearch(item, keyword, searchSelectData.value, searchValue.value);
    }

    // 没有选中过滤标签
    if (!item) {
      // 过滤掉已经选过的标签
      const selected = (searchValue.value || []).map(value => value.id);
      return searchSelectData.value.filter(item => !selected.includes(item.id));
    }

    // 远程加载执行人
    if (item.id === 'creator') {
      if (!keyword) {
        return [];
      }
      return getUserList({
        fuzzy_lookups: keyword,
      }).then(res => res.results.map(item => ({
        id: item.username,
        name: item.username,
      })));
    }

    // 不需要远层加载
    return searchSelectData.value.find(set => set.id === item.id)?.children || [];
  };

  const fetchData = (loading?:boolean) => {
    const params = getSearchSelectorParams(searchValue.value);
    tableRef.value!.fetchData(params, { ...sortValue }, loading);
    isInit.value = false;
  };

  const handleCopy = <T,>(dataList: T[], field: keyof T) => {
    const copyList = dataList.reduce((prevList, tableItem) => {
      const value = String(tableItem[field]);
      if (value && value !== '--' && !prevList.includes(value)) {
        prevList.push(value);
      }
      return prevList;
    }, [] as string[]);
    copy(copyList.join('\n'));
  }

  // 获取列表数据下的实例子列表
  const getInstanceListByRole = (dataList: TendbhaModel[], field: keyof TendbhaModel) => dataList.reduce((result, curRow) => {
    result.push(...curRow[field] as TendbhaModel['masters']);
    return result;
  }, [] as TendbhaModel['masters']);

  const handleCopySelected = <T,>(field: keyof T, role?: keyof TendbhaModel) => {
    if(role) {
      handleCopy(getInstanceListByRole(selected.value, role) as T[], field)
      return;
    }
    handleCopy(selected.value as T[], field)
  }

  const handleCopyAll = async <T,>(field: keyof T, role?: keyof TendbhaModel) => {
    const allData = await tableRef.value!.getAllData<TendbhaModel>();
    if(allData.length === 0) {
      Message({
        theme: 'primary',
        message: t('暂无数据可复制'),
      });
      return;
    }
    if(role) {
      handleCopy(getInstanceListByRole(allData, role) as T[], field)
      return;
    }
    handleCopy(allData as T[], field)
  }

  const handleSelection = (data: TendbhaModel, list: TendbhaModel[]) => {
    selected.value = list;
    selectedClusterList.value = list;
  };

  const handleShowAuthorize = (selected: TendbhaModel[] = []) => {
    authorizeState.isShow = true;
    authorizeState.selected = selected;
  };

  const handleShowCreateSubscribeRuleSlider = (data?: ColumnData['data']) => {
    if (data) {
      // 单个集群订阅
      selectedClusterList.value = [data];
    }
    showCreateSubscribeRuleSlider.value = true;
  };

  const handleShowDataExportSlider = (data: ColumnData['data']) => {
    currentData.value = data
    showDataExportSlider.value = true;
  };

  const handleClearSelected = () => {
    selected.value = [];
    authorizeState.selected = [];
  };

  // excel 授权
  const handleShowExcelAuthorize = () => {
    isShowExcelAuthorize.value = true;
  };

  /**
   * 查看详情
   */
  const handleToDetails = (id: number) => {
    stretchLayoutSplitScreen();
    clusterId.value = id;
  };

  const handleGoWebconsole = (clusterId: number) => {
    router.push({
      name: 'MySQLWebconsole',
      query: {
        clusterId
      }
    });
  }

  /**
   * 申请实例
   */
  const handleApply = () => {
    router.push({
      name: 'SelfServiceApplyHa',
      query: {
        bizId: globalBizsStore.currentBizId,
        from: route.name as string,
      },
    });
  };

  const handleBatchOperationSuccess = () => {
    tableRef.value!.clearSelected();
    fetchData();
  }

  onMounted(() => {
    if (route.query.id && !clusterId.value) {
      handleToDetails(Number(route.query.id));
    }
  });
</script>

<style lang="less" scoped>
  @import '@styles/mixins.less';

  .mysql-ha-cluster-list-page {
    height: 100%;
    padding: 24px 0;
    margin: 0 24px;
    overflow: hidden;

    .operation-box {
      display: flex;
      flex-wrap: wrap;
      margin-bottom: 16px;

      .bk-search-select {
        flex: 1;
        max-width: 500px;
        min-width: 320px;
        margin-left: auto;
      }
    }

    .table-wrapper {
      background-color: white;
    }

    :deep(td .vxe-cell) {
      .domain {
        display: flex;
        flex-wrap: wrap;

        .bk-search-select {
          flex: 1;
          max-width: 320px;
          min-width: 320px;
          margin-left: auto;
        }
      }

      .slave-entry {
        line-height: 22px;
      }

      .is-stand-by {
        color: #531dab !important;
        background: #f9f0ff !important;
      }

      .db-icon-copy,
      .db-icon-visible1 {
        display: none;
        margin-top: 1px;
        margin-left: 4px;
        color: @primary-color;
        cursor: pointer;
      }

      :deep(.cluster-name-container) {
        display: flex;
        align-items: center;
        padding: 8px 0;
        overflow: hidden;

        .cluster-name {
          line-height: 16px;

          &__alias {
            color: @light-gray;
          }
        }

        .cluster-tags {
          display: flex;
          margin-left: 4px;
          align-items: center;
          flex-wrap: wrap;
        }

        .cluster-tag {
          margin: 2px 0;
          flex-shrink: 0;
        }
      }
    }

    :deep(th:hover) {
      .db-icon-copy {
        display: inline-block !important;
      }
    }

    :deep(td:hover) {
      .db-icon-copy,
      .db-icon-visible1 {
        display: inline-block !important;
      }
    }

    :deep(.is-offline) {
      a {
        color: @gray-color;
      }

      .vxe-cell {
        color: @disable-color;
      }
    }

    :deep(.operations-more) {
      .db-icon-more {
        font-size: 16px;
        color: @default-color;
        cursor: pointer;

        &:hover {
          background-color: @bg-disable;
          border-radius: 2px;
        }
      }
    }
  }
</style>
