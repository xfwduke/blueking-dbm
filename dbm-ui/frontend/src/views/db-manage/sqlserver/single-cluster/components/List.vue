<template>
  <div class="sqlserver-single-cluster-list">
    <div class="header-action mb-16">
      <div>
        <BkButton
          v-db-console="'sqlserver.singleClusterList.instanceApply'"
          theme="primary"
          @click="handleApply">
          {{ t('申请实例') }}
        </BkButton>
        <ClusterBatchOperation
          v-db-console="'sqlserver.singleClusterList.batchOperation'"
          class="ml-8"
          :cluster-type="ClusterTypes.SQLSERVER_SINGLE"
          :selected="selected"
          @success="handleBatchOperationSuccess" />
        <BkButton
          v-db-console="'sqlserver.singleClusterList.importAuthorize'"
          class="ml-8"
          @click="handleShowExcelAuthorize">
          {{ t('导入授权') }}
        </BkButton>
        <DropdownExportExcel
          v-db-console="'sqlserver.singleClusterList.export'"
          export-type="cluster"
          :has-selected="hasSelected"
          :ids="selectedIds"
          type="sqlserver_single" />
        <ClusterIpCopy
          v-db-console="'sqlserver.singleClusterList.batchCopy'"
          :selected="selected" />
      </div>
      <DbSearchSelect
        class="header-select"
        :data="searchSelectData"
        :get-menu-list="getMenuList"
        :model-value="searchValue"
        :placeholder="t('请输入或选择条件搜索')"
        unique-select
        :validate-values="validateSearchValues"
        @change="handleSearchValueChange" />
    </div>
    <div class="table-wrapper">
      <DbTable
        ref="tableRef"
        :columns="columns"
        :data-source="getSingleClusterList"
        releate-url-query
        :row-class="setRowClass"
        selectable
        :settings="settings"
        :show-overflow="false"
        show-overflow-tips
        @clear-search="clearSearchValue"
        @column-filter="columnFilterChange"
        @column-sort="columnSortChange"
        @selection="handleSelection"
        @setting-change="updateTableSettings" />
    </div>
  </div>
  <!-- 集群授权 -->
  <ClusterAuthorize
    v-model="authorizeShow"
    :account-type="AccountTypes.SQLSERVER"
    :cluster-types="[ClusterTypes.SQLSERVER_SINGLE]"
    :selected="authorizeSelected"
    @success="handleClearSelected" />
  <!-- excel 导入授权 -->
  <ExcelAuthorize
    v-model:is-show="isShowExcelAuthorize"
    :cluster-type="ClusterTypes.SQLSERVER_SINGLE"
    :ticket-type="TicketTypes.SQLSERVER_EXCEL_AUTHORIZE_RULES" />
  <ClusterReset
    v-if="currentData"
    v-model:is-show="isShowClusterReset"
    :data="currentData" />
</template>

<script setup lang="tsx">
  import { Message } from 'bkui-vue';
  import type { ISearchItem } from 'bkui-vue/lib/search-select/utils';
  import { useI18n } from 'vue-i18n';
  import {
    useRoute,
    useRouter,
  } from 'vue-router';

  import SqlServerSingleModel from '@services/model/sqlserver/sqlserver-single';
  import {
    getSingleClusterList,
    getSqlServerInstanceList,
  } from '@services/source/sqlserverSingleCluster';
  import { getUserList } from '@services/source/user';

  import {
    useCopy,
    useLinkQueryColumnSerach,
    useStretchLayout,
    useTableSettings,
  } from '@hooks';

  import { useGlobalBizs } from '@stores';

  import {
    AccountTypes,
    ClusterTypes,
    DBTypes,
    TicketTypes,
    UserPersonalSettings,
  } from '@common/const';

  import RenderClusterStatus from '@components/cluster-status/Index.vue';
  import DbTable from '@components/db-table/index.vue';
  import TextOverflowLayout from '@components/text-overflow-layout/Index.vue';

  import ClusterAuthorize from '@views/db-manage/common/cluster-authorize/Index.vue';
  import ClusterBatchOperation from '@views/db-manage/common/cluster-batch-opration/Index.vue'
  import ClusterCapacityUsageRate from '@views/db-manage/common/cluster-capacity-usage-rate/Index.vue'
  import EditEntryConfig from '@views/db-manage/common/cluster-entry-config/Index.vue';
  import ClusterIpCopy from '@views/db-manage/common/cluster-ip-copy/Index.vue';
  import DropdownExportExcel from '@views/db-manage/common/dropdown-export-excel/index.vue';
  import ExcelAuthorize from '@views/db-manage/common/ExcelAuthorize.vue';
  import { useOperateClusterBasic } from '@views/db-manage/common/hooks';
  import OperationBtnStatusTips from '@views/db-manage/common/OperationBtnStatusTips.vue';
  import RenderCellCopy from '@views/db-manage/common/render-cell-copy/Index.vue';
  import RenderHeadCopy from '@views/db-manage/common/render-head-copy/Index.vue';
  import RenderInstances from '@views/db-manage/common/render-instances/RenderInstances.vue';
  import RenderOperationTag from '@views/db-manage/common/RenderOperationTagNew.vue';
  import ClusterReset from '@views/db-manage/sqlserver/components/cluster-reset/Index.vue'

  import {
    getMenuListSearch,
    getSearchSelectorParams,
    // isRecentDays,
  } from '@utils';

  const singleClusterData = defineModel<{ clusterId: number }>('singleClusterData');

  const router = useRouter();
  const route = useRoute();
  const { currentBizId } = useGlobalBizs();
  const copy = useCopy();

  const {
    t,
    locale,
  } = useI18n();

  const { handleDisableCluster, handleEnableCluster, handleDeleteCluster } = useOperateClusterBasic(
    ClusterTypes.SQLSERVER,
    {
      onSuccess: () => fetchData(),
    },
  );

  const {
    isOpen: isStretchLayoutOpen,
    splitScreen: stretchLayoutSplitScreen,
    handleOpenChange,
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
    searchType: ClusterTypes.SQLSERVER_SINGLE,
    attrs: [
      'bk_cloud_id',
      'db_module_id',
      'major_version',
      'region',
      'time_zone',
    ],
    fetchDataFn: () => fetchData(isInit),
    defaultSearchItem: {
      name: t('访问入口'),
      id: 'domain',
    }
  });

  const tableRef = ref<InstanceType<typeof DbTable>>();
  const isShowExcelAuthorize = ref(false);
  const isShowClusterReset = ref(false)
  const currentData = ref<SqlServerSingleModel>()
  const selected = ref<SqlServerSingleModel[]>([])

  /** 集群授权 */
  const authorizeShow = ref(false);

  const authorizeSelected = ref<{
    master_domain: string,
    cluster_name: string,
    db_module_name: string,
  }[]>([]);

  const hasSelected = computed(() => selected.value.length > 0);
  const selectedIds = computed(() => selected.value.map(item => item.id));
  const isCN = computed(() => locale.value === 'zh-cn');

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
      multiple: true,
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
      name: t('模块'),
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
      return isCN.value ? 180 : 200;
    }
    return 100;
  });

  const columns = computed(() => [
    {
      label: 'ID',
      field: 'id',
      fixed: 'left',
      width: 100,
    },
    {
      label: t('访问入口'),
      field: 'master_domain',
      fixed: 'left',
      minWidth: 320,
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
          {t('访问入口')}
        </RenderHeadCopy>
      ),
      render: ({ data }: { data: SqlServerSingleModel }) => (
        <TextOverflowLayout>
          {{
            default: () => (
              <auth-button
                action-id="sqlserver_view"
                permission={data.permission.sqlserver_view}
                resource-id={data.id}
                text
                theme="primary"
                onClick={() => handleToDetails(data)}>
                {data.master_domain}
              </auth-button>
            ),
            append: () => (
              <>
                {
                  data.operationTagTips.map(item => (
                    <RenderOperationTag
                      class="cluster-tag"
                      data={item} />
                  ))
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
                <span v-db-console="sqlserver.singleClusterList.modifyEntryConfiguration">
                  <EditEntryConfig
                    id={data.id}
                    bizId={data.bk_biz_id}
                    permission={data.permission.access_entry_edit}
                    resource={DBTypes.SQLSERVER}
                    onSuccess={fetchData} />
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
      width: 200,
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
      render: ({ data }: { data: SqlServerSingleModel }) => (
        <TextOverflowLayout>
          {{
            default: () => data.cluster_name,
            append: () => (
              <>
                <db-icon
                  v-bk-tooltips={t('复制集群名称')}
                  type="copy"
                  onClick={() => copy(data.cluster_name)} />
              </>
            ),
          }}
        </TextOverflowLayout>
      ),
    },
    {
      label: t('管控区域'),
      field: 'bk_cloud_id',
      filter: {
        list: columnAttrs.value.bk_cloud_id,
        checked: columnCheckedMap.value.bk_cloud_id,
      },
      width: 90,
      render: ({ data }: { data: SqlServerSingleModel }) => <span>{data.bk_cloud_name || '--'}</span>,
    },
    {
      label: t('状态'),
      field: 'status',
      minWidth: 100,
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
      render: ({ data }: { data: SqlServerSingleModel }) => <RenderClusterStatus data={data.status} />,
    },
    {
      label: t('容量使用率'),
      field: 'cluster_stats',
      width: 240,
      showOverflowTooltip: false,
      render: ({ data }: { data: SqlServerSingleModel }) => <ClusterCapacityUsageRate clusterStats={data.cluster_stats} />
    },
    {
      label: t('实例'),
      field: 'storages',
      width: 180,
      minWidth: 180,
      showOverflowTooltip: false,
      renderHead: () => (
        <RenderHeadCopy
          hasSelected={hasSelected.value}
          onHandleCopySelected={(field) => handleCopySelected(field, 'storages')}
          onHandleCopyAll={(field) => handleCopyAll(field, 'storages')}
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
          {t('实例')}
        </RenderHeadCopy>
      ),
      render: ({ data }: { data: SqlServerSingleModel }) => (
        <RenderInstances
          highlightIps={batchSearchIpInatanceList.value}
          data={data.storages}
          dataSource={getSqlServerInstanceList}
          title={t('【inst】实例预览', { inst: data.bk_cloud_name })}
          role="storages"
          clusterId={data.id}
        />
      )
    },
    {
      label: t('所属DB模块'),
      field: 'db_module_id',
      width: 140,
      showOverflowTooltip: true,
      filter: {
        list: columnAttrs.value.db_module_id,
        checked: columnCheckedMap.value.db_module_id,
      },
      render: ({ data }: { data: SqlServerSingleModel }) => <span>{data.db_module_name || '--'}</span>,
    },
    {
      label: t('版本'),
      field: 'major_version',
      minWidth: 180,
      width: 180,
      filter: {
        list: columnAttrs.value.major_version,
        checked: columnCheckedMap.value.major_version,
      },
      render: ({ data }: { data: SqlServerSingleModel }) => <span>{data.major_version || '--'}</span>,
    },
    {
      label: t('地域'),
      field: 'region',
      minWidth: 100,
      width: 100,
      filter: {
        list: columnAttrs.value.region,
        checked: columnCheckedMap.value.region,
      },
      render: ({ data }: { data: SqlServerSingleModel }) => <span>{data.region || '--'}</span>,
    },
    {
      label: t('创建人'),
      field: 'creator',
      width: 140,
      render: ({ data }: { data: SqlServerSingleModel }) => <span>{data.creator || '--'}</span>,
    },
    {
      label: t('部署时间'),
      field: 'create_at',
      width: 160,
      sort: true,
      render: ({ data }: { data: SqlServerSingleModel }) => <span>{data.createAtDisplay || '--'}</span>,
    },
    {
      label: t('时区'),
      field: 'cluster_time_zone',
      width: 100,
      filter: {
        list: columnAttrs.value.time_zone,
        checked: columnCheckedMap.value.time_zone,
      },
      render: ({ data }: { data: SqlServerSingleModel }) => <span>{data.cluster_time_zone || '--'}</span>,
    },
    {
      label: t('操作'),
      field: '',
      width: tableOperationWidth.value,
      fixed: isStretchLayoutOpen.value ? false : 'right',
      render: ({ data }: { data: SqlServerSingleModel }) => {
        const oprations = []

        if (data.isOnline) {
          oprations.push([
            <bk-button
              v-db-console="sqlserver.singleClusterList.authorize"
              text
              theme="primary"
              onClick={ () => handleShowAuthorize([data]) }>
              { t('授权') }
            </bk-button>,
            <OperationBtnStatusTips
              data={ data }
              v-db-console="sqlserver.singleClusterList.disable">
              <bk-button
                text
                theme="primary"
                class="ml-16"
                disabled={data.operationDisabled}
                onClick={ () => handleDisableCluster([data]) }>
                { t('禁用') }
              </bk-button>
            </OperationBtnStatusTips>
          ])
        } else {
          oprations.push([
            <OperationBtnStatusTips
              data={ data }
              v-db-console="sqlserver.singleClusterList.enable">
              <bk-button
                text
                theme="primary"
                disabled={data.isStarting}
                onClick={ () => handleEnableCluster([data]) }>
                { t('启用') }
              </bk-button>
            </OperationBtnStatusTips>,
            <OperationBtnStatusTips
              data={ data }
              v-db-console="sqlserver.singleClusterList.reset">
              <bk-button
                text
                theme="primary"
                class="ml-16"
                disabled={Boolean(data.operationTicketId)}
                onClick={() => handleResetCluster(data)}>
                { t('重置') }
              </bk-button>
            </OperationBtnStatusTips>
          ])
        }

        oprations.push(
          <OperationBtnStatusTips
            data={ data }
            v-db-console="sqlserver.singleClusterList.delete">
            <bk-button
              v-bk-tooltips={{
                disabled: data.isOffline,
                content: t('请先禁用集群')
              }}
              text
              theme="primary"
              class="ml-16"
              disabled={data.isOnline || Boolean(data.operationTicketId)}
              onClick={ () => handleDeleteCluster([data]) }>
              { t('删除') }
            </bk-button>
          </OperationBtnStatusTips>
        )

        return oprations
      }

    },
  ]);

  // 设置用户个人表头信息
  const defaultSettings = {
    fields: (columns.value || []).filter(item => item.field).map(item => ({
      label: item.label,
      field: item.field ,
      disabled: ['master_domain'].includes(item.field as string),
    })),
    checked: [
      'master_domain',
      'status',
      'cluster_stats',
      'storages',
      'db_module_id',
      'major_version',
      'region',
    ],
    showLineHeight: false,
    trigger: 'manual' as const,
  };

  const {
    settings,
    updateTableSettings,
  } = useTableSettings(UserPersonalSettings.SQLSERVER_SINGLE_TABLE_SETTINGS, defaultSettings);

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

  const handleResetCluster = (data: SqlServerSingleModel) => {
    currentData.value = data
    isShowClusterReset.value = true
  }

  // excel 授权
  const handleShowExcelAuthorize = () => {
    isShowExcelAuthorize.value = true;
  };

  let isInit = true;
  const fetchData = (loading?: boolean) => {
    tableRef.value!.fetchData(
      { ...getSearchSelectorParams(searchValue.value) },
      { bk_biz_id: window.PROJECT_CONFIG.BIZ_ID, ...sortValue },
      loading
    );
    isInit = false;
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
  const getInstanceListByRole = (dataList: SqlServerSingleModel[], field: keyof SqlServerSingleModel) => dataList.reduce((result, curRow) => {
    result.push(...curRow[field] as SqlServerSingleModel['storages']);
    return result;
  }, [] as SqlServerSingleModel['storages']);

  const handleCopySelected = <T,>(field: keyof T, role?: keyof SqlServerSingleModel) => {
    if(role) {
      handleCopy(getInstanceListByRole(selected.value, role) as T[], field)
      return;
    }
    handleCopy(selected.value as T[], field)
  }

  const handleCopyAll = async <T,>(field: keyof T, role?: keyof SqlServerSingleModel) => {
    const allData = await tableRef.value!.getAllData<SqlServerSingleModel>();
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

  // 设置行样式
  const setRowClass = (row: SqlServerSingleModel) => {
    const classStack = [];
    if (row.isNew) {
      classStack.push('is-new-row');
    }
    if (singleClusterData.value && row.id === singleClusterData.value.clusterId) {
      classStack.push('is-selected-row');
    }
    return classStack.join(' ');
  };

  const handleSelection = (key: number[], list: Record<number, SqlServerSingleModel>[]) => {
    selected.value = list as unknown as SqlServerSingleModel[];
  };

  const handleClearSelected = () => {
    selected.value = [];
    authorizeSelected.value = [];
  };

  const handleShowAuthorize = (selected: {
    master_domain: string,
    cluster_name: string,
    db_module_name: string,
  }[]) => {
    authorizeShow.value = true;
    authorizeSelected.value = selected;
  };

  /**
   * 查看详情
   */
  const handleToDetails = (
    data: SqlServerSingleModel,
    isAllSpread: boolean = false,
  ) => {
    stretchLayoutSplitScreen();
    singleClusterData.value = { clusterId: data.id };
    if (isAllSpread) {
      handleOpenChange('left');
    }
  };

  /**
   * 申请实例
   */
  const handleApply = () => {
    router.push({
      name: 'SqlServiceSingleApply',
      query: {
        bizId: currentBizId,
        from: String(route.name),
      },
    });
  };

  const handleBatchOperationSuccess = () => {
    tableRef.value!.clearSelected();
    fetchData();
  }
</script>
<style lang="less">
  @import '@styles/mixins.less';

  .sqlserver-single-cluster-list {
    height: 100%;
    padding: 24px 0;
    margin: 0 24px;
    overflow: hidden;

    .header-action {
      display: flex;
      flex-wrap: wrap;

      .header-select {
        flex: 1;
        max-width: 500px;
        min-width: 320px;
        margin-left: auto;
      }
    }

    td .vxe-cell {
      .db-icon-copy,
      .db-icon-link,
      .db-icon-visible1 {
        display: none;
        margin-left: 4px;
        color: @primary-color;
        cursor: pointer;
      }

      .operations-more {
        .db-icon-more {
          display: block;
          font-size: @font-size-normal;
          font-weight: bold;
          color: @default-color;
          cursor: pointer;

          &:hover {
            background-color: @bg-disable;
            border-radius: 2px;
          }
        }
      }
    }

    td:hover {
      .db-icon-copy,
      .db-icon-link,
      .db-icon-visible1 {
        display: inline-block !important;
      }
    }
  }
</style>
