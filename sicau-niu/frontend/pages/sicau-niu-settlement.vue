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
  CollegeRankItem,
  DashboardData,
  DeviceCluster,
  FeedRankItem,
  RuleConfig,
} from "./settlement-client";

import { computed, onMounted, ref } from "vue";

import {
  Button,
  Card,
  InputNumber,
  Input as AInput,
  message,
  Progress,
  Statistic,
  Table,
  Tag,
} from "ant-design-vue";

import { useAccess } from "@vben/access";

import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import {
  createArchive,
  exportPlayers,
  getActivity,
  getCollegeRanking,
  getDashboard,
  getFeedRanking,
  getFriendRanking,
  getRules,
  getRiskAnomalies,
  getRiskDeviceClusters,
  issueCertificates,
  listArchives,
  updateRules,
} from "./settlement-client";

const { hasAccessByCodes } = useAccess();

const pluginAccessCodes = {
  export: "sicau-niu:settlement:export",
  issue: "sicau-niu:settlement:issue",
  archive: "sicau-niu:settlement:archive",
  rules: "sicau-niu:settlement:rules",
} as const;

function defaultRuleConfig(): RuleConfig {
  return {
    activationLbsThresholdMeters: 50,
    posterCampusBadge: "",
    checkinMinAmount: 20,
    checkinMaxAmount: 50,
    stealDailyTargets: 12,
    stealDailyLimit: 5,
    stealMinAmount: 5,
    stealMaxAmount: 20,
    giftDailyLimit: 12,
    giftMinAmount: 12,
    ironBonusThresholdMeters: 12,
    rankingTopN: 100,
    anomalyFeedDailyThreshold: 100,
    anomalyStealDailyThreshold: 5,
    anomalyListLimit: 200,
    miniappUrl: "",
  };
}

const dashboard = ref<DashboardData | null>(null);
const activity = ref<ActivityData | null>(null);
const clusters = ref<DeviceCluster[]>([]);
const anomalies = ref<AnomalyAlert[]>([]);
const archives = ref<ArchiveItem[]>([]);
const ruleForm = ref<RuleConfig>(defaultRuleConfig());
const feedRanking = ref<FeedRankItem[]>([]);
const friendRanking = ref<FeedRankItem[]>([]);
const collegeRanking = ref<CollegeRankItem[]>([]);

const exporting = ref(false);
const issuing = ref(false);
const archiving = ref(false);
const savingRules = ref(false);
const issueHonorId = ref<number | undefined>(undefined);
const archiveTitle = ref("");

const operationMetricCards = [
  { key: "firstActivatorCount", label: "首发英雄", accent: "gold" },
  { key: "feedingCount", label: "喂草次数", accent: "green" },
  { key: "stealCount", label: "偷草次数", accent: "red" },
  { key: "giftCount", label: "送草次数", accent: "blue" },
  { key: "checkinCount", label: "签到次数", accent: "cyan" },
] as const;

const activationProgress = computed(() => {
  const total = metricValue("totalNiuCount");
  if (total <= 0) {
    return 0;
  }
  return Number(
    Math.min(100, (metricValue("activatedNiuCount") / total) * 100).toFixed(1),
  );
});

const latestActiveUsers = computed(() => {
  const rows = activity.value?.dau ?? [];
  return rows.length > 0 ? rows[rows.length - 1]?.activeUsers ?? 0 : 0;
});

const riskIssueCount = computed(
  () => clusters.value.length + anomalies.value.length,
);
const topFeedPlayer = computed(() => feedRanking.value[0]);
const topCollege = computed(() => collegeRanking.value[0]);

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
  {
    title: "活跃玩家",
    dataIndex: "activeUsers",
    key: "activeUsers",
    width: 120,
  },
];

const playerRankingColumns = [
  { title: "排名", dataIndex: "rank", key: "rank", width: 70 },
  { title: "玩家", key: "nickname" },
  { title: "喂草效果", dataIndex: "total", key: "total", width: 110 },
];

const collegeRankingColumns = [
  { title: "排名", dataIndex: "rank", key: "rank", width: 70 },
  { title: "院系", key: "collegeName" },
  { title: "喂草效果", dataIndex: "total", key: "total", width: 110 },
];

async function loadDashboard() {
  dashboard.value = await getDashboard();
}

async function loadActivity() {
  activity.value = await getActivity(14);
}

async function loadRules() {
  ruleForm.value = await getRules();
}

async function loadRankings() {
  const [feed, friend, college] = await Promise.all([
    getFeedRanking(),
    getFriendRanking(),
    getCollegeRanking(),
  ]);
  feedRanking.value = feed;
  friendRanking.value = friend;
  collegeRanking.value = college;
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

function validateRules(): boolean {
  if (ruleForm.value.checkinMaxAmount < ruleForm.value.checkinMinAmount) {
    message.warning("签到草量上限不能小于下限");
    return false;
  }
  if (ruleForm.value.stealMaxAmount < ruleForm.value.stealMinAmount) {
    message.warning("偷草草量上限不能小于下限");
    return false;
  }
  return true;
}

function metricValue(key: string): number {
  const data = dashboard.value as Record<string, number> | null;
  return data ? data[key] ?? 0 : 0;
}

function feedRankName(record: FeedRankItem | undefined): string {
  if (!record) {
    return "暂无数据";
  }
  return record.nickname || `#${record.userId}`;
}

function collegeRankName(record: CollegeRankItem | undefined): string {
  if (!record) {
    return "暂无数据";
  }
  return record.collegeName || `#${record.collegeId}`;
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

async function onSaveRules() {
  if (!validateRules()) {
    return;
  }
  savingRules.value = true;
  try {
    ruleForm.value = await updateRules(ruleForm.value);
    message.success("规则配置已保存");
    await Promise.all([loadRankings(), loadAnomalies()]);
  } finally {
    savingRules.value = false;
  }
}

onMounted(() => {
  loadDashboard();
  loadActivity();
  loadRules();
  loadRankings();
  loadClusters();
  loadAnomalies();
  loadArchives();
});
</script>

<template>
  <Page>
    <div class="settlement-workbench">
      <section class="settlement-hero" data-testid="settlement-dashboard">
        <div class="hero-main">
          <div class="hero-copy">
            <div class="hero-eyebrow">运营结算</div>
            <h2 class="hero-title">寻牛活动运营概览</h2>
            <p class="hero-desc">
              汇总激活、互动、榜单、风控和归档状态，支撑活动日常复盘与结算动作。
            </p>
          </div>
          <div class="hero-status">
            <Tag color="processing">榜单 Top {{ ruleForm.rankingTopN }}</Tag>
            <Tag :color="riskIssueCount > 0 ? 'error' : 'success'">
              风险 {{ riskIssueCount }}
            </Tag>
            <Tag color="default">近 14 天活跃</Tag>
          </div>
        </div>

        <div class="hero-grid">
          <div class="hero-metric hero-metric--primary">
            <span class="metric-label">参与玩家</span>
            <strong class="metric-value">{{
              metricValue("playerCount")
            }}</strong>
            <span class="metric-foot">当前活动报名与互动用户</span>
          </div>
          <div class="hero-metric">
            <span class="metric-label">牛只激活</span>
            <strong class="metric-value">{{ activationProgress }}%</strong>
            <Progress
              :percent="activationProgress"
              :show-info="false"
              size="small"
              status="active"
            />
            <span class="metric-foot">
              {{ metricValue("activatedNiuCount") }} /
              {{ metricValue("totalNiuCount") }}
            </span>
          </div>
          <div class="hero-metric hero-metric--warm">
            <span class="metric-label">喂草总效果</span>
            <strong class="metric-value">{{
              metricValue("feedTotalEffect")
            }}</strong>
            <span class="metric-foot">今日活跃 {{ latestActiveUsers }} 人</span>
          </div>
          <div class="hero-metric hero-metric--blue">
            <span class="metric-label">已发证书</span>
            <strong class="metric-value">
              {{ metricValue("certificateGrantedCount") }}
            </strong>
            <span class="metric-foot">可继续按荣誉批量发放</span>
          </div>
        </div>
      </section>

      <div class="main-grid">
        <div class="left-stack">
          <Card class="section-card" data-testid="settlement-activity">
            <template #title>
              <div class="card-title">
                <span>活跃度（日活 / 留存）</span>
                <small>沉淀最近 14 天日活与关键留存表现</small>
              </div>
            </template>

            <div class="activity-layout">
              <div class="retention-column">
                <div class="retention-grid">
                  <div class="retention-box">
                    <Statistic
                      title="次日留存"
                      :value="retentionText(activity?.retentionD1?.rate)"
                    />
                    <div class="retention-detail">
                      cohort {{ activity?.retentionD1?.cohortUsers ?? 0 }} /
                      回访
                      {{ activity?.retentionD1?.returnedUsers ?? 0 }}
                    </div>
                  </div>
                  <div class="retention-box">
                    <Statistic
                      title="7 日留存"
                      :value="retentionText(activity?.retentionD7?.rate)"
                    />
                    <div class="retention-detail">
                      cohort {{ activity?.retentionD7?.cohortUsers ?? 0 }} /
                      回访
                      {{ activity?.retentionD7?.returnedUsers ?? 0 }}
                    </div>
                  </div>
                </div>
                <div class="latest-dau">
                  <span>最新日活</span>
                  <strong>{{ latestActiveUsers }}</strong>
                </div>
              </div>

              <div class="table-panel">
                <div class="panel-head">
                  <div class="panel-title">
                    <strong>近 14 天日活</strong>
                    <small>按天观察活动回流情况</small>
                  </div>
                </div>
                <Table
                  :data-source="activity?.dau ?? []"
                  :columns="dauColumns"
                  :pagination="false"
                  row-key="date"
                  size="small"
                  :scroll="{ y: 220 }"
                />
              </div>
            </div>
          </Card>

          <Card class="section-card" data-testid="settlement-rules">
            <template #title>
              <div class="card-title">
                <span>互动规则配置</span>
                <small>按业务场景分组维护，保存后立即影响运行规则</small>
              </div>
            </template>
            <template #extra>
              <Button
                v-if="hasAccessByCodes([pluginAccessCodes.rules])"
                type="primary"
                :loading="savingRules"
                data-testid="settlement-rules-save"
                @click="onSaveRules"
              >
                保存规则
              </Button>
            </template>

            <div class="rules-overview">
              <div class="rules-overview-item">
                <span>LBS 判距</span>
                <strong>{{ ruleForm.activationLbsThresholdMeters }}m</strong>
              </div>
              <div class="rules-overview-item">
                <span>排行榜范围</span>
                <strong>Top {{ ruleForm.rankingTopN }}</strong>
              </div>
              <div class="rules-overview-item">
                <span>异常告警上限</span>
                <strong>{{ ruleForm.anomalyListLimit }}</strong>
              </div>
            </div>

            <div class="rule-group-grid">
              <section class="rule-group">
                <div class="rule-group-head">
                  <h4>激活与展示</h4>
                  <p>控制 LBS 激活、铁牛加成与海报展示信息。</p>
                </div>
                <div class="field-grid">
                  <label class="field">
                    <span class="field-label">LBS 判距(米)</span>
                    <InputNumber
                      v-model:value="ruleForm.activationLbsThresholdMeters"
                      :min="1"
                      style="width: 100%"
                    />
                  </label>
                  <label class="field">
                    <span class="field-label">铁牛加成阈值(米)</span>
                    <InputNumber
                      v-model:value="ruleForm.ironBonusThresholdMeters"
                      :min="1"
                      style="width: 100%"
                    />
                  </label>
                  <label class="field full-field">
                    <span class="field-label">海报标识</span>
                    <AInput v-model:value="ruleForm.posterCampusBadge" />
                  </label>
                  <label class="field full-field">
                    <span class="field-label">小程序回跳地址</span>
                    <AInput v-model:value="ruleForm.miniappUrl" />
                  </label>
                </div>
              </section>

              <section class="rule-group">
                <div class="rule-group-head">
                  <h4>签到与送草</h4>
                  <p>约束每日签到收益、送草次数和单次下限。</p>
                </div>
                <div class="field-grid">
                  <label class="field">
                    <span class="field-label">签到草量下限</span>
                    <InputNumber
                      v-model:value="ruleForm.checkinMinAmount"
                      :min="1"
                      style="width: 100%"
                    />
                  </label>
                  <label class="field">
                    <span class="field-label">签到草量上限</span>
                    <InputNumber
                      v-model:value="ruleForm.checkinMaxAmount"
                      :min="1"
                      style="width: 100%"
                    />
                  </label>
                  <label class="field">
                    <span class="field-label">每日送草上限</span>
                    <InputNumber
                      v-model:value="ruleForm.giftDailyLimit"
                      :min="1"
                      style="width: 100%"
                    />
                  </label>
                  <label class="field">
                    <span class="field-label">单次送草下限</span>
                    <InputNumber
                      v-model:value="ruleForm.giftMinAmount"
                      :min="1"
                      style="width: 100%"
                    />
                  </label>
                </div>
              </section>

              <section class="rule-group">
                <div class="rule-group-head">
                  <h4>偷草策略</h4>
                  <p>控制候选名单、每日次数和单次草量区间。</p>
                </div>
                <div class="field-grid">
                  <label class="field">
                    <span class="field-label">偷草名单人数</span>
                    <InputNumber
                      v-model:value="ruleForm.stealDailyTargets"
                      :min="1"
                      style="width: 100%"
                    />
                  </label>
                  <label class="field">
                    <span class="field-label">每日偷草上限</span>
                    <InputNumber
                      v-model:value="ruleForm.stealDailyLimit"
                      :min="1"
                      style="width: 100%"
                    />
                  </label>
                  <label class="field">
                    <span class="field-label">单次偷草下限</span>
                    <InputNumber
                      v-model:value="ruleForm.stealMinAmount"
                      :min="1"
                      style="width: 100%"
                    />
                  </label>
                  <label class="field">
                    <span class="field-label">单次偷草上限</span>
                    <InputNumber
                      v-model:value="ruleForm.stealMaxAmount"
                      :min="1"
                      style="width: 100%"
                    />
                  </label>
                </div>
              </section>

              <section class="rule-group">
                <div class="rule-group-head">
                  <h4>榜单与风控</h4>
                  <p>控制榜单规模与异常喂草、异常偷草判定阈值。</p>
                </div>
                <div class="field-grid">
                  <label class="field">
                    <span class="field-label">排行榜 TopN</span>
                    <InputNumber
                      v-model:value="ruleForm.rankingTopN"
                      :min="1"
                      :max="500"
                      style="width: 100%"
                    />
                  </label>
                  <label class="field">
                    <span class="field-label">异常告警上限</span>
                    <InputNumber
                      v-model:value="ruleForm.anomalyListLimit"
                      :min="1"
                      :max="500"
                      style="width: 100%"
                    />
                  </label>
                  <label class="field">
                    <span class="field-label">喂草异常阈值</span>
                    <InputNumber
                      v-model:value="ruleForm.anomalyFeedDailyThreshold"
                      :min="1"
                      style="width: 100%"
                    />
                  </label>
                  <label class="field">
                    <span class="field-label">偷草异常阈值</span>
                    <InputNumber
                      v-model:value="ruleForm.anomalyStealDailyThreshold"
                      :min="1"
                      style="width: 100%"
                    />
                  </label>
                </div>
              </section>
            </div>
          </Card>
        </div>

        <div class="right-stack">
          <Card class="section-card">
            <template #title>
              <div class="card-title">
                <span>关键互动</span>
                <small>用于快速判断活动热度和互动结构</small>
              </div>
            </template>

            <div class="metric-grid">
              <div
                v-for="card in operationMetricCards"
                :key="card.key"
                class="metric-tile"
                :class="`metric-tile--${card.accent}`"
              >
                <span>{{ card.label }}</span>
                <strong>{{ metricValue(card.key) }}</strong>
              </div>
            </div>
            <div class="total-tile">
              <span>全部牛只</span>
              <strong>{{ metricValue("totalNiuCount") }}</strong>
            </div>
          </Card>

          <Card class="section-card action-card">
            <template #title>
              <div class="card-title">
                <span>运营动作</span>
                <small>导出、发证和结算归档集中处理</small>
              </div>
            </template>

            <div class="action-list">
              <div
                v-if="hasAccessByCodes([pluginAccessCodes.export])"
                class="action-item"
              >
                <div class="action-copy">
                  <span>玩家名册</span>
                  <small>导出结算所需用户、学院、激活和喂草数据。</small>
                </div>
                <Button
                  type="primary"
                  :loading="exporting"
                  data-testid="settlement-export"
                  @click="onExport"
                >
                  导出玩家名册
                </Button>
              </div>

              <div
                v-if="hasAccessByCodes([pluginAccessCodes.issue])"
                class="action-item"
              >
                <div class="action-copy">
                  <span>批量发证</span>
                  <small>按荣誉 ID 向达标玩家发放电子证书。</small>
                </div>
                <div class="inline-controls">
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
                </div>
              </div>

              <div
                v-if="hasAccessByCodes([pluginAccessCodes.archive])"
                class="action-item"
              >
                <div class="action-copy">
                  <span>结算归档</span>
                  <small>保存当前运营结算快照，便于后续公示复核。</small>
                </div>
                <div class="inline-controls archive-controls">
                  <AInput
                    v-model:value="archiveTitle"
                    placeholder="归档标题"
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
                </div>
              </div>
            </div>
          </Card>
        </div>
      </div>

      <Card class="section-card" data-testid="settlement-rankings">
        <template #title>
          <div class="card-title">
            <span>排行榜数据</span>
            <small>面向运营侧的喂草榜、好友榜和院系榜对比视图</small>
          </div>
        </template>
        <template #extra>
          <Tag color="processing">Top {{ ruleForm.rankingTopN }}</Tag>
        </template>

        <div class="rank-summary">
          <div class="rank-summary-item">
            <span>喂草榜首</span>
            <strong>{{ feedRankName(topFeedPlayer) }}</strong>
          </div>
          <div class="rank-summary-item">
            <span>好友榜记录</span>
            <strong>{{ friendRanking.length }}</strong>
          </div>
          <div class="rank-summary-item">
            <span>领先院系</span>
            <strong>{{ collegeRankName(topCollege) }}</strong>
          </div>
        </div>

        <div class="rank-grid">
          <div class="rank-panel">
            <div class="panel-head">
              <div class="panel-title">
                <strong>喂草榜</strong>
                <small>全量玩家喂草效果排名</small>
              </div>
            </div>
            <Table
              :data-source="feedRanking"
              :columns="playerRankingColumns"
              :pagination="false"
              row-key="rank"
              size="small"
              :scroll="{ y: 260 }"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'nickname'">
                  {{ feedRankName(record as FeedRankItem) }}
                </template>
              </template>
            </Table>
          </div>

          <div class="rank-panel">
            <div class="panel-head">
              <div class="panel-title">
                <strong>好友榜</strong>
                <small>好友互动范围内的喂草效果排名</small>
              </div>
            </div>
            <Table
              :data-source="friendRanking"
              :columns="playerRankingColumns"
              :pagination="false"
              row-key="rank"
              size="small"
              :scroll="{ y: 260 }"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'nickname'">
                  {{ feedRankName(record as FeedRankItem) }}
                </template>
              </template>
            </Table>
          </div>

          <div class="rank-panel">
            <div class="panel-head">
              <div class="panel-title">
                <strong>院系榜</strong>
                <small>按院系统计喂草总效果</small>
              </div>
            </div>
            <Table
              :data-source="collegeRanking"
              :columns="collegeRankingColumns"
              :pagination="false"
              row-key="rank"
              size="small"
              :scroll="{ y: 260 }"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'collegeName'">
                  {{ collegeRankName(record as CollegeRankItem) }}
                </template>
              </template>
            </Table>
          </div>
        </div>
      </Card>

      <div class="risk-grid">
        <Card class="section-card">
          <template #title>
            <div class="card-title">
              <span>风控告警 · 一机多号</span>
              <small>同设备指纹下的多账号聚合</small>
            </div>
          </template>
          <template #extra>
            <Tag :color="clusters.length > 0 ? 'warning' : 'success'">
              {{ clusters.length }} 组
            </Tag>
          </template>
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

        <Card class="section-card">
          <template #title>
            <div class="card-title">
              <span>风控告警 · 异常喂草/偷草</span>
              <small>按规则阈值识别异常互动用户</small>
            </div>
          </template>
          <template #extra>
            <Tag :color="anomalies.length > 0 ? 'error' : 'success'">
              {{ anomalies.length }} 条
            </Tag>
          </template>
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
      </div>

      <Card class="section-card">
        <template #title>
          <div class="card-title">
            <span>结算归档</span>
            <small>历史结算快照与公示依据</small>
          </div>
        </template>
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
    </div>
  </Page>
</template>

<style scoped>
.settlement-workbench {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.settlement-hero {
  padding: 20px;
  overflow: hidden;
  background: var(--ant-color-bg-container, #fff);
  border: 1px solid var(--ant-color-border-secondary, #edf0f3);
  border-radius: 8px;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 6%);
}

.hero-main {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  justify-content: space-between;
}

.hero-copy {
  min-width: 0;
}

.hero-eyebrow {
  font-size: 12px;
  font-weight: 600;
  line-height: 1.3;
  color: var(--ant-color-primary, #1677ff);
}

.hero-title {
  margin: 4px 0 6px;
  font-size: 24px;
  font-weight: 700;
  line-height: 1.22;
  color: var(--ant-color-text, #1f2937);
}

.hero-desc {
  max-width: 720px;
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  color: var(--ant-color-text-secondary, #667085);
}

.hero-status {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.hero-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-top: 18px;
}

.hero-metric {
  display: flex;
  min-height: 120px;
  padding: 16px;
  flex-direction: column;
  justify-content: space-between;
  background: #f8fafc;
  border: 1px solid #edf0f3;
  border-radius: 8px;
}

.hero-metric--primary {
  background: #f6fbf7;
  border-color: #d9eadc;
}

.hero-metric--warm {
  background: #fffaf0;
  border-color: #f3dfaf;
}

.hero-metric--blue {
  background: #f3f8ff;
  border-color: #d7e6fb;
}

.metric-label,
.metric-foot {
  font-size: 12px;
  line-height: 1.4;
  color: var(--ant-color-text-secondary, #667085);
}

.metric-value {
  margin: 8px 0 6px;
  font-size: 30px;
  font-weight: 700;
  line-height: 1.1;
  color: var(--ant-color-text, #1f2937);
}

.main-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(320px, 0.65fr);
  gap: 16px;
  align-items: start;
}

.left-stack,
.right-stack {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.section-card {
  border-radius: 8px;
}

.section-card :deep(.ant-card-head) {
  min-height: 56px;
  border-bottom-color: var(--ant-color-border-secondary, #edf0f3);
}

.section-card :deep(.ant-card-body) {
  padding: 16px;
}

.card-title {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 2px;
}

.card-title span {
  font-weight: 600;
  line-height: 1.35;
}

.card-title small {
  overflow: hidden;
  font-size: 12px;
  font-weight: 400;
  line-height: 1.45;
  color: var(--ant-color-text-secondary, #667085);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.activity-layout {
  display: grid;
  grid-template-columns: minmax(220px, 0.85fr) minmax(0, 1.15fr);
  gap: 16px;
}

.retention-column {
  min-width: 0;
}

.retention-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.retention-box,
.rules-overview-item,
.rank-summary-item,
.total-tile {
  padding: 14px;
  background: #fafbfc;
  border: 1px solid #edf0f3;
  border-radius: 8px;
}

.retention-detail {
  margin-top: 8px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--ant-color-text-secondary, #667085);
}

.latest-dau {
  display: flex;
  padding: 12px 14px;
  margin-top: 12px;
  align-items: center;
  justify-content: space-between;
  background: #f3f8ff;
  border: 1px solid #d7e6fb;
  border-radius: 8px;
}

.latest-dau span {
  font-size: 12px;
  color: var(--ant-color-text-secondary, #667085);
}

.latest-dau strong {
  font-size: 22px;
  color: var(--ant-color-text, #1f2937);
}

.table-panel,
.rank-panel {
  min-width: 0;
  overflow: hidden;
  border: 1px solid #edf0f3;
  border-radius: 8px;
}

.panel-head {
  display: flex;
  padding: 12px;
  align-items: flex-start;
  justify-content: space-between;
  background: #f8fafc;
  border-bottom: 1px solid #edf0f3;
}

.panel-title {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 2px;
}

.panel-title strong {
  font-size: 13px;
  line-height: 1.4;
  color: var(--ant-color-text, #1f2937);
}

.panel-title small {
  font-size: 12px;
  line-height: 1.4;
  color: var(--ant-color-text-secondary, #667085);
}

.rules-overview,
.rank-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 14px;
}

.rules-overview-item,
.rank-summary-item,
.total-tile,
.metric-tile {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
}

.rules-overview-item span,
.rank-summary-item span,
.total-tile span,
.metric-tile span {
  overflow: hidden;
  font-size: 12px;
  line-height: 1.35;
  color: var(--ant-color-text-secondary, #667085);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rules-overview-item strong,
.rank-summary-item strong,
.total-tile strong,
.metric-tile strong {
  overflow: hidden;
  font-size: 18px;
  line-height: 1.25;
  color: var(--ant-color-text, #1f2937);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rule-group-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.rule-group {
  min-width: 0;
  padding: 14px;
  background: var(--ant-color-bg-container, #fff);
  border: 1px solid #edf0f3;
  border-radius: 8px;
}

.rule-group-head h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.4;
  color: var(--ant-color-text, #1f2937);
}

.rule-group-head p {
  margin: 4px 0 12px;
  font-size: 12px;
  line-height: 1.55;
  color: var(--ant-color-text-secondary, #667085);
}

.field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.field {
  min-width: 0;
}

.field-label {
  display: block;
  margin-bottom: 6px;
  overflow: hidden;
  font-size: 12px;
  line-height: 1.35;
  color: var(--ant-color-text-secondary, #667085);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.full-field {
  grid-column: 1 / -1;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.metric-tile {
  min-height: 82px;
  padding: 12px;
  background: #fafbfc;
  border: 1px solid #edf0f3;
  border-left: 3px solid #8c8c8c;
  border-radius: 8px;
}

.metric-tile--gold {
  border-left-color: #d48806;
}

.metric-tile--green {
  border-left-color: #389e0d;
}

.metric-tile--red {
  border-left-color: #cf1322;
}

.metric-tile--blue {
  border-left-color: #1677ff;
}

.metric-tile--cyan {
  border-left-color: #08979c;
}

.total-tile {
  margin-top: 10px;
  background: #f6fbf7;
  border-color: #d9eadc;
}

.action-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.action-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  flex-direction: column;
  align-items: stretch;
  background: #fafbfc;
  border: 1px solid #edf0f3;
  border-radius: 8px;
}

.action-copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 3px;
}

.action-copy span {
  font-size: 13px;
  font-weight: 600;
  line-height: 1.35;
  color: var(--ant-color-text, #1f2937);
}

.action-copy small {
  font-size: 12px;
  line-height: 1.45;
  color: var(--ant-color-text-secondary, #667085);
  overflow-wrap: anywhere;
}

.inline-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-start;
}

.inline-controls :deep(.ant-input-number) {
  width: 140px;
}

.archive-controls {
  display: grid;
  grid-template-columns: minmax(160px, 1fr) auto;
  width: 100%;
  min-width: 0;
}

.archive-controls :deep(.ant-input) {
  width: 100%;
}

.rank-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.risk-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

@media (max-width: 1280px) {
  .hero-grid,
  .rank-grid,
  .risk-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .main-grid,
  .activity-layout {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .settlement-hero {
    padding: 16px;
  }

  .hero-main {
    flex-direction: column;
  }

  .hero-status {
    justify-content: flex-start;
  }

  .hero-title {
    font-size: 20px;
  }

  .hero-grid,
  .retention-grid,
  .rules-overview,
  .rule-group-grid,
  .field-grid,
  .metric-grid,
  .rank-summary,
  .rank-grid,
  .risk-grid {
    grid-template-columns: 1fr;
  }

  .card-title small {
    white-space: normal;
  }

  .action-item {
    grid-template-columns: 1fr;
  }

  .inline-controls {
    justify-content: flex-start;
  }

  .archive-controls {
    min-width: 0;
  }

  .archive-controls :deep(.ant-input) {
    width: 100%;
  }
}
</style>
