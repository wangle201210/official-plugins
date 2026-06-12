<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-record-activation",
  title: "激活记录",
};
</script>

<script setup lang="ts">
import type { ActivationAttemptRecord, ActivationRecord } from "./record-client";

import { ref } from "vue";

import { IconifyIcon } from "@vben/icons";

import { Image, ImagePreviewGroup, Space, TabPane, Tabs, Tag } from "ant-design-vue";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import { listActivationAttempts, listActivations } from "./record-client";

const previewVisible = ref(false);
const previewImage = ref("");
const activeTab = ref("activations");

const attemptResultOptions = [
  { label: "成功", value: "success" },
  { label: "附近无牛", value: "no_nearby" },
  { label: "超出判距", value: "out_of_range" },
  { label: "速度异常", value: "speed_anomaly" },
];

const attemptResultLabels: Record<string, string> = {
  no_nearby: "附近无牛",
  out_of_range: "超出判距",
  speed_anomaly: "速度异常",
  success: "成功",
};

const attemptResultColors: Record<string, string> = {
  no_nearby: "default",
  out_of_range: "warning",
  speed_anomaly: "error",
  success: "success",
};

const [Grid] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "InputNumber",
        fieldName: "userId",
        label: "玩家ID",
        componentProps: { min: 1, class: "w-full" },
      },
      {
        component: "InputNumber",
        fieldName: "niuId",
        label: "牛ID",
        componentProps: { min: 1, class: "w-full" },
      },
    ],
    commonConfig: { labelWidth: 70, componentProps: { allowClear: true } },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      { field: "nickname", title: "玩家", minWidth: 140 },
      { field: "niuName", title: "牛名称", minWidth: 120 },
      { field: "niuCode", title: "牛编码", width: 120 },
      { field: "activityDate", title: "激活日期", width: 130 },
      {
        field: "isFirst",
        title: "首发",
        width: 80,
        formatter: ({ cellValue }) => (cellValue === 1 ? "是" : "否"),
      },
      { field: "orderNo", title: "到场顺序", width: 100 },
      {
        field: "photoPath",
        slots: { default: "photo" },
        title: "照片",
        width: 110,
      },
      {
        field: "activatedAt",
        title: "激活时间",
        width: 180,
        formatter: ({ cellValue }) => formatTimestamp(cellValue),
      },
    ],
    height: "auto",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues = {},
        ) => {
          return await listActivations({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: { keyField: "id" },
    id: "sicau-niu-record-activation-grid",
  },
});

const [AttemptGrid] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "InputNumber",
        fieldName: "userId",
        label: "玩家ID",
        componentProps: { min: 1, class: "w-full" },
      },
      {
        component: "InputNumber",
        fieldName: "niuId",
        label: "牛ID",
        componentProps: { min: 1, class: "w-full", placeholder: "激活或最近牛ID" },
      },
      {
        component: "Select",
        fieldName: "result",
        label: "结果",
        componentProps: {
          allowClear: true,
          options: attemptResultOptions,
          class: "w-full",
        },
      },
    ],
    commonConfig: { labelWidth: 70, componentProps: { allowClear: true } },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      { field: "nickname", title: "玩家", minWidth: 140 },
      {
        field: "result",
        slots: { default: "attemptResult" },
        title: "结果",
        width: 110,
      },
      {
        field: "nearestNiuName",
        slots: { default: "nearestNiu" },
        title: "最近牛",
        minWidth: 150,
      },
      {
        field: "location",
        slots: { default: "attemptLocation" },
        title: "上报坐标",
        minWidth: 180,
      },
      {
        field: "distanceM",
        title: "距离(米)",
        width: 110,
        formatter: ({ cellValue }) => formatMeters(cellValue),
      },
      {
        field: "thresholdM",
        title: "判距(米)",
        width: 110,
        formatter: ({ cellValue }) => formatMeters(cellValue),
      },
      {
        field: "photoPath",
        slots: { default: "attemptPhoto" },
        title: "照片",
        width: 110,
      },
      {
        field: "attemptedAt",
        title: "打卡时间",
        width: 180,
        formatter: ({ cellValue }) => formatTimestamp(cellValue),
      },
    ],
    height: "auto",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues = {},
        ) => {
          return await listActivationAttempts({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: { keyField: "id" },
    id: "sicau-niu-record-activation-attempt-grid",
  },
});

function openPhotoPreview(row: ActivationAttemptRecord | ActivationRecord) {
  const photoPath = row.photoPath?.trim();
  if (!photoPath) {
    return;
  }
  previewImage.value = photoPath;
  previewVisible.value = true;
}

function handlePreviewVisibleChange(visible: boolean) {
  previewVisible.value = visible;
  if (!visible) {
    previewImage.value = "";
  }
}

function attemptResultLabel(result: string) {
  return attemptResultLabels[result] ?? result;
}

function attemptResultColor(result: string) {
  return attemptResultColors[result] ?? "default";
}

function formatNiu(name: string, code: string, id: number) {
  if (name && code) {
    return `${name}(${code})`;
  }
  if (name) {
    return name;
  }
  if (code) {
    return code;
  }
  return id > 0 ? `#${id}` : "-";
}

function formatCoordinate(value: number) {
  return Number.isFinite(value) ? value.toFixed(6) : "-";
}

function formatMeters(value: number) {
  return Number.isFinite(value) && value > 0 ? value.toFixed(1) : "-";
}
</script>

<template>
  <Page :auto-content-height="true" content-class="flex min-h-0 flex-col">
    <Tabs
      v-model:active-key="activeTab"
      :animated="false"
      class="activation-record-tabs flex min-h-0 flex-1 flex-col overflow-hidden"
    >
      <TabPane key="activations" tab="正式激活">
        <div class="activation-record-tab-pane">
          <Grid class="min-h-0 flex-1 overflow-hidden" table-title="激活记录">
            <template #photo="{ row }">
              <Space>
                <ghost-button
                  v-if="row.photoPath"
                  :data-testid="`sicau-niu-activation-photo-${row.id}`"
                  @click.stop="openPhotoPreview(row)"
                >
                  <IconifyIcon icon="ant-design:eye-outlined" class="mr-1" />
                  查看
                </ghost-button>
                <span v-else>-</span>
              </Space>
            </template>
          </Grid>
        </div>
      </TabPane>
      <TabPane key="attempts" tab="打卡尝试">
        <div class="activation-record-tab-pane">
          <AttemptGrid
            class="min-h-0 flex-1 overflow-hidden"
            table-title="打卡尝试记录"
          >
            <template #attemptResult="{ row }">
              <Tag :color="attemptResultColor(row.result)">
                {{ attemptResultLabel(row.result) }}
              </Tag>
            </template>
            <template #nearestNiu="{ row }">
              {{
                formatNiu(row.nearestNiuName, row.nearestNiuCode, row.nearestNiuId)
              }}
            </template>
            <template #attemptLocation="{ row }">
              {{ formatCoordinate(row.lat) }}, {{ formatCoordinate(row.lng) }}
            </template>
            <template #attemptPhoto="{ row }">
              <Space>
                <ghost-button
                  v-if="row.photoPath"
                  :data-testid="`sicau-niu-activation-attempt-photo-${row.id}`"
                  @click.stop="openPhotoPreview(row)"
                >
                  <IconifyIcon icon="ant-design:eye-outlined" class="mr-1" />
                  查看
                </ghost-button>
                <span v-else>-</span>
              </Space>
            </template>
          </AttemptGrid>
        </div>
      </TabPane>
    </Tabs>
    <ImagePreviewGroup
      :preview="{
        visible: previewVisible,
        onVisibleChange: handlePreviewVisibleChange,
      }"
    >
      <Image
        class="hidden"
        :src="previewImage"
        data-testid="sicau-niu-activation-photo-preview"
      />
    </ImagePreviewGroup>
  </Page>
</template>

<style scoped>
.activation-record-tabs :deep(.ant-tabs-content-holder) {
  flex: 1 1 auto;
  height: 100%;
  max-height: 100%;
  min-height: 0;
  overflow: hidden;
}

.activation-record-tabs :deep(.ant-tabs-content),
.activation-record-tabs :deep(.ant-tabs-tabpane),
.activation-record-tabs :deep(.ant-tabs-tabpane-active) {
  display: flex;
  flex: 1 1 auto;
  height: 100%;
  max-height: 100%;
  min-height: 0;
  overflow: hidden;
}

.activation-record-tab-pane {
  display: flex;
  flex: 1 1 auto;
  height: 100%;
  max-height: 100%;
  min-height: 0;
  overflow: hidden;
}
</style>
