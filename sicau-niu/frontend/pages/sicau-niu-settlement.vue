<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-settlement",
  title: "运营结算",
};
</script>

<script setup lang="ts">
import type {
  ActivityData,
  AnomalyAlert,
  ArchiveItem,
  DashboardData,
  DeviceCluster,
} from "./settlement-client";

import { onMounted, ref } from "vue";

import {
  Button,
  Card,
  Col,
  InputNumber,
  Input as AInput,
  message,
  Row,
  Space,
  Statistic,
  Table,
} from "ant-design-vue";

import { useAccess } from "@vben/access";

import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import {
  createArchive,
  exportPlayers,
  getActivity,
  getDashboard,
  getRiskAnomalies,
  getRiskDeviceClusters,
  issueCertificates,
  listArchives,
} from "./settlement-client";

const { hasAccessByCodes } = useAccess();

const pluginAccessCodes = {
  export: "sicau-niu:settlement:export",
  issue: "sicau-niu:settlement:issue",
  archive: "sicau-niu:settlement:archive",
} as const;

const dashboard = ref<DashboardData | null>(null);
const activity = ref<ActivityData | null>(null);
const clusters = ref<DeviceCluster[]>([]);
const anomalies = ref<AnomalyAlert[]>([]);
const archives = ref<ArchiveItem[]>([]);

const exporting = ref(false);
const issuing = ref(false);
const archiving = ref(false);
const issueHonorId = ref<number | undefined>(undefined);
const archiveTitle = ref("");

const metricCards = [
  { key: "playerCount", label: "参与玩家" },
  { key: "activatedNiuCount", label: "已激活牛" },
  { key: "totalNiuCount", label: "全部牛只" },
  { key: "firstActivatorCount", label: "首发英雄" },
  { key: "feedingCount", label: "喂草次数" },
  { key: "feedTotalEffect", label: "喂草总效果" },
  { key: "stealCount", label: "偷草次数" },
  { key: "giftCount", label: "送草次数" },
  { key: "checkinCount", label: "签到次数" },
  { key: "certificateGrantedCount", label: "已发证书" },
] as const;

const clusterColumns = [
  { title: "设备指纹", dataIndex: "fingerprint", key: "fingerprint" },
  { title: "账号数", dataIndex: "count", key: "count", width: 100 },
  { title: "成员昵称", key: "members" },
];

const archiveColumns = [
  { title: "ID", dataIndex: "id", key: "id", width: 80 },
  { title: "标题", dataIndex: "title", key: "title" },
  { title: "归档时间", key: "archivedAt", width: 200 },
];

const dauColumns = [
  { title: "日期", dataIndex: "date", key: "date" },
  { title: "活跃玩家", dataIndex: "activeUsers", key: "activeUsers", width: 120 },
];

async function loadDashboard() {
  dashboard.value = await getDashboard();
}

async function loadActivity() {
  activity.value = await getActivity(14);
}

function retentionText(rate: number | undefined): string {
  if (rate === undefined) {
    return "—";
  }
  return `${(rate * 100).toFixed(1)}%`;
}

async function loadClusters() {
  clusters.value = await getRiskDeviceClusters();
}

const anomalyTypeLabels: Record<string, string> = {
  feed: "喂草",
  steal: "偷草",
};

const anomalyColumns = [
  { title: "玩家", dataIndex: "nickname", key: "nickname" },
  { title: "类型", key: "type", width: 90 },
  { title: "日期", dataIndex: "date", key: "date", width: 130 },
  { title: "当日次数", dataIndex: "count", key: "count", width: 100 },
  { title: "阈值", dataIndex: "threshold", key: "threshold", width: 90 },
];

async function loadAnomalies() {
  anomalies.value = await getRiskAnomalies();
}

async function loadArchives() {
  archives.value = await listArchives();
}

function metricValue(key: string): number {
  const data = dashboard.value as Record<string, number> | null;
  return data ? (data[key] ?? 0) : 0;
}

function membersText(cluster: DeviceCluster): string {
  return cluster.members.map((m) => m.nickname || `#${m.userId}`).join("、");
}

async function onExport() {
  exporting.value = true;
  try {
    const data = await exportPlayers();
    const header = [
      "userId",
      "nickname",
      "identityType",
      "collegeName",
      "grade",
      "activationCount",
      "feedTotalEffect",
    ];
    const lines = [header.join(",")];
    for (const row of data.list) {
      lines.push(
        [
          row.userId,
          row.nickname,
          row.identityType,
          row.collegeName,
          row.grade,
          row.activationCount,
          row.feedTotalEffect,
        ]
          .map((v) => `"${String(v ?? "").replace(/"/g, '""')}"`)
          .join(","),
      );
    }
    const blob = new Blob([`﻿${lines.join("\n")}`], {
      type: "text/csv;charset=utf-8;",
    });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "sicau-niu-players.csv";
    anchor.click();
    URL.revokeObjectURL(url);
    if (data.truncated) {
      message.warning(`名册超过导出上限,已截断;总玩家数 ${data.total}`);
    } else {
      message.success(`已导出 ${data.list.length} 名玩家`);
    }
  } finally {
    exporting.value = false;
  }
}

async function onIssue() {
  if (!issueHonorId.value || issueHonorId.value <= 0) {
    message.warning("请输入要发放的证书荣誉 ID");
    return;
  }
  issuing.value = true;
  try {
    const result = await issueCertificates(issueHonorId.value);
    message.success(
      `达标 ${result.eligible} 人,新发 ${result.issued} 人,跳过 ${result.skipped} 人`,
    );
    await loadDashboard();
  } finally {
    issuing.value = false;
  }
}

async function onArchive() {
  if (!archiveTitle.value.trim()) {
    message.warning("请输入归档标题");
    return;
  }
  archiving.value = true;
  try {
    await createArchive(archiveTitle.value.trim());
    message.success("结算归档已创建");
    archiveTitle.value = "";
    await loadArchives();
  } finally {
    archiving.value = false;
  }
}

onMounted(() => {
  loadDashboard();
  loadActivity();
  loadClusters();
  loadAnomalies();
  loadArchives();
});
</script>

<template>
  <Page>
    <Card title="数据看板" class="mb-4" data-testid="settlement-dashboard">
      <Row :gutter="[16, 16]">
        <Col
          v-for="card in metricCards"
          :key="card.key"
          :xs="12"
          :sm="8"
          :md="6"
          :lg="4"
        >
          <Statistic :title="card.label" :value="metricValue(card.key)" />
        </Col>
      </Row>
    </Card>

    <Card title="活跃度(日活 / 留存)" class="mb-4" data-testid="settlement-activity">
      <Row :gutter="16">
        <Col :xs="24" :md="8">
          <Statistic
            title="次日留存"
            :value="retentionText(activity?.retentionD1?.rate)"
          />
          <div class="text-xs text-gray-400">
            cohort {{ activity?.retentionD1?.cohortUsers ?? 0 }} / 回访
            {{ activity?.retentionD1?.returnedUsers ?? 0 }}
          </div>
        </Col>
        <Col :xs="24" :md="8">
          <Statistic
            title="7 日留存"
            :value="retentionText(activity?.retentionD7?.rate)"
          />
          <div class="text-xs text-gray-400">
            cohort {{ activity?.retentionD7?.cohortUsers ?? 0 }} / 回访
            {{ activity?.retentionD7?.returnedUsers ?? 0 }}
          </div>
        </Col>
        <Col :xs="24" :md="8">
          <div class="mb-2 text-sm">近 14 天日活</div>
          <Table
            :data-source="activity?.dau ?? []"
            :columns="dauColumns"
            :pagination="false"
            row-key="date"
            size="small"
            :scroll="{ y: 180 }"
          />
        </Col>
      </Row>
    </Card>

    <Card title="运营动作" class="mb-4">
      <Space :size="24" wrap>
        <Button
          v-if="hasAccessByCodes([pluginAccessCodes.export])"
          type="primary"
          :loading="exporting"
          data-testid="settlement-export"
          @click="onExport"
        >
          导出玩家名册
        </Button>

        <span
          v-if="hasAccessByCodes([pluginAccessCodes.issue])"
          class="inline-flex items-center gap-2"
        >
          <InputNumber
            v-model:value="issueHonorId"
            :min="1"
            placeholder="证书荣誉 ID"
            data-testid="settlement-issue-id"
          />
          <Button
            :loading="issuing"
            data-testid="settlement-issue"
            @click="onIssue"
          >
            批量发证
          </Button>
        </span>

        <span
          v-if="hasAccessByCodes([pluginAccessCodes.archive])"
          class="inline-flex items-center gap-2"
        >
          <AInput
            v-model:value="archiveTitle"
            placeholder="归档标题"
            style="width: 220px"
            data-testid="settlement-archive-title"
          />
          <Button
            type="primary"
            :loading="archiving"
            data-testid="settlement-archive"
            @click="onArchive"
          >
            创建结算归档
          </Button>
        </span>
      </Space>
    </Card>

    <Card title="风控告警 · 一机多号" class="mb-4">
      <Table
        :data-source="clusters"
        :columns="clusterColumns"
        :pagination="false"
        row-key="fingerprint"
        size="small"
        data-testid="settlement-risk-table"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'members'">
            {{ membersText(record as DeviceCluster) }}
          </template>
        </template>
      </Table>
    </Card>

    <Card title="风控告警 · 异常喂草/偷草" class="mb-4">
      <Table
        :data-source="anomalies"
        :columns="anomalyColumns"
        :pagination="false"
        :row-key="(r) => `${r.userId}-${r.type}-${r.date}`"
        size="small"
        data-testid="settlement-anomaly-table"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'type'">
            {{
              anomalyTypeLabels[(record as AnomalyAlert).type] ??
              (record as AnomalyAlert).type
            }}
          </template>
        </template>
      </Table>
    </Card>

    <Card title="结算归档">
      <Table
        :data-source="archives"
        :columns="archiveColumns"
        :pagination="false"
        row-key="id"
        size="small"
        data-testid="settlement-archive-table"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'archivedAt'">
            {{ formatTimestamp((record as ArchiveItem).archivedAt) }}
          </template>
        </template>
      </Table>
    </Card>
  </Page>
</template>
