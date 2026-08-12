<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-transport",
  title: "云搬牛管理",
};
</script>

<script setup lang="ts">
import type { TransportTeamItem } from "./transport-client";

import { onMounted, ref } from "vue";

import { useAccess } from "@vben/access";
import { useVbenModal } from "@vben/common-ui";

import { Col, Row, Space, Statistic, Tabs } from "ant-design-vue";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { DictTag } from "#/components/dict";
import { Page } from "#/plugins/dynamic";
import { useDictStore } from "#/store/dict";
import { formatTimestamp } from "#/utils/time";

import TransportNameModal from "./components/transport-name-modal.vue";
import {
  getTransportStats,
  listTransportReports,
  listTransportTeams,
  type TransportStats,
} from "./transport-client";

const { hasAccessByCodes } = useAccess();
const dictStore = useDictStore();
const activeTab = ref("teams");
const teamStatusDicts = ref<any[]>([]);
const TabPane = Tabs.TabPane;
const stats = ref<TransportStats>({
  effectiveTeamCount: 0,
  invalidTeamCount: 0,
  activeMemberCount: 0,
  reportCount: 0,
  contributionMeters: 0,
});

const canUpdate = () => hasAccessByCodes(["sicau-niu:transport:update"]);
const canAudit = () => hasAccessByCodes(["sicau-niu:transport:audit"]);

const [NameModal, nameModalApi] = useVbenModal({
  connectedComponent: TransportNameModal,
});

const [TeamGrid, teamGridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      { component: "Input", fieldName: "name", label: "团名称" },
      {
        component: "Select",
        componentProps: {
          options: [],
        },
        fieldName: "status",
        label: "状态",
      },
    ],
    commonConfig: { labelWidth: 70, componentProps: { allowClear: true } },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      { field: "name", title: "团名称", minWidth: 180 },
      {
        field: "status",
        slots: { default: "status" },
        title: "状态",
        width: 100,
      },
      { field: "creatorName", title: "创建人", minWidth: 140 },
      { field: "memberCount", title: "当前成员", width: 100 },
      {
        field: "totalContributionMeters",
        formatter: ({ cellValue }) => formatMeters(cellValue),
        title: "累计贡献",
        width: 130,
      },
      {
        field: "lastActiveAt",
        formatter: ({ cellValue }) => formatTimestamp(cellValue),
        title: "最近活跃",
        width: 180,
      },
      {
        field: "createdAt",
        formatter: ({ cellValue }) => formatTimestamp(cellValue),
        title: "创建时间",
        width: 180,
      },
      {
        field: "action",
        fixed: "right",
        slots: { default: "action" },
        title: "操作",
        width: 100,
      },
    ],
    height: "auto",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          values = {},
        ) =>
          await listTransportTeams({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...values,
          }),
      },
    },
    rowConfig: { keyField: "id" },
    id: "sicau-niu-transport-team-grid",
  },
});

const [ReportGrid] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "InputNumber",
        componentProps: { min: 1, class: "w-full" },
        fieldName: "teamId",
        label: "团ID",
      },
      {
        component: "InputNumber",
        componentProps: { min: 1, class: "w-full" },
        fieldName: "userId",
        label: "玩家ID",
      },
      {
        component: "DatePicker",
        componentProps: { valueFormat: "YYYY-MM-DD", class: "w-full" },
        fieldName: "activityDate",
        label: "日期",
      },
    ],
    commonConfig: { labelWidth: 70, componentProps: { allowClear: true } },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      { field: "teamName", title: "团名称", minWidth: 170 },
      { field: "userName", title: "玩家", minWidth: 140 },
      { field: "activityDate", title: "日期", width: 120 },
      {
        field: "start",
        slots: { default: "start" },
        title: "起点",
        minWidth: 190,
      },
      {
        field: "end",
        slots: { default: "end" },
        title: "本次点位",
        minWidth: 190,
      },
      {
        field: "contributionMeters",
        formatter: ({ cellValue }) => formatMeters(cellValue),
        title: "本次贡献",
        width: 120,
      },
      {
        field: "acceptedAt",
        formatter: ({ cellValue }) => formatTimestamp(cellValue),
        title: "接收时间",
        width: 180,
      },
    ],
    height: "auto",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          values = {},
        ) =>
          await listTransportReports({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...values,
          }),
      },
    },
    rowConfig: { keyField: "id" },
    id: "sicau-niu-transport-report-grid",
  },
});

onMounted(async () => {
  [teamStatusDicts.value] = await Promise.all([
    dictStore.getDictOptionsAsync("sicau_niu_transport_team_status"),
    loadStats(),
  ]);
  teamGridApi.formApi.updateSchema([
    {
      fieldName: "status",
      componentProps: {
        options: teamStatusDicts.value.map((item: any) => ({
          label: item.label,
          value: item.value,
        })),
      },
    },
  ]);
});

async function loadStats() {
  stats.value = await getTransportStats();
}

function editName(row: TransportTeamItem) {
  nameModalApi.setData({ id: row.id, name: row.name });
  nameModalApi.open();
}

async function reloadTeams() {
  await Promise.all([teamGridApi.query(), loadStats()]);
}

function formatMeters(value: unknown) {
  const meters = Number(value ?? 0);
  return `${meters.toLocaleString("zh-CN")} 米`;
}

function formatCoordinate(lat?: number, lng?: number) {
  if (lat === undefined || lng === undefined) {
    return "首次上报";
  }
  return `${lat.toFixed(6)}, ${lng.toFixed(6)}`;
}
</script>

<template>
  <Page :auto-content-height="true" content-class="flex min-h-0 flex-col">
    <Row :gutter="12" class="mb-4" data-testid="sicau-niu-transport-stats">
      <Col :xs="12" :md="8" :xl="4"
        ><Statistic title="有效团" :value="stats.effectiveTeamCount"
      /></Col>
      <Col :xs="12" :md="8" :xl="4"
        ><Statistic title="失效团" :value="stats.invalidTeamCount"
      /></Col>
      <Col :xs="12" :md="8" :xl="4"
        ><Statistic title="当前成员" :value="stats.activeMemberCount"
      /></Col>
      <Col :xs="12" :md="8" :xl="4"
        ><Statistic title="成功上报" :value="stats.reportCount"
      /></Col>
      <Col :xs="24" :md="8" :xl="8"
        ><Statistic title="累计贡献（米）" :value="stats.contributionMeters"
      /></Col>
    </Row>

    <Tabs
      v-model:active-key="activeTab"
      :animated="false"
      class="transport-tabs flex min-h-0 flex-1 flex-col overflow-hidden"
      destroy-inactive-tab-pane
    >
      <TabPane key="teams" tab="团管理">
        <div class="transport-tab-pane">
          <TeamGrid
            class="min-h-0 flex-1 overflow-hidden"
            table-title="云搬牛团"
          >
            <template #status="{ row }">
              <DictTag :dicts="teamStatusDicts" :value="String(row.status)" />
            </template>
            <template #action="{ row }">
              <Space>
                <ghost-button
                  v-if="canUpdate()"
                  :data-testid="`sicau-niu-transport-rename-${row.id}`"
                  @click.stop="editName(row)"
                >
                  改名
                </ghost-button>
              </Space>
            </template>
          </TeamGrid>
        </div>
      </TabPane>
      <TabPane v-if="canAudit()" key="reports" tab="位置上报审计">
        <div class="transport-tab-pane">
          <ReportGrid
            class="min-h-0 flex-1 overflow-hidden"
            table-title="成功上报事实"
          >
            <template #start="{ row }">{{
              formatCoordinate(row.startLat, row.startLng)
            }}</template>
            <template #end="{ row }">{{
              formatCoordinate(row.endLat, row.endLng)
            }}</template>
          </ReportGrid>
        </div>
      </TabPane>
    </Tabs>
    <NameModal @reload="reloadTeams" />
  </Page>
</template>

<style scoped>
.transport-tabs :deep(.ant-tabs-content-holder) {
  flex: 1 1 auto;
  height: 100%;
  max-height: 100%;
  min-height: 0;
  overflow: hidden;
}

.transport-tabs :deep(.ant-tabs-content),
.transport-tabs :deep(.ant-tabs-tabpane),
.transport-tabs :deep(.ant-tabs-tabpane-active) {
  display: flex;
  flex: 1 1 auto;
  height: 100%;
  max-height: 100%;
  min-height: 0;
  overflow: hidden;
}

.transport-tab-pane {
  display: flex;
  flex: 1 1 auto;
  height: 100%;
  max-height: 100%;
  min-height: 0;
  overflow: hidden;
}
</style>
