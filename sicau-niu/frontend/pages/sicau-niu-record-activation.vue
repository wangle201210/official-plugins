<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-record-activation",
  title: "激活记录",
};
</script>

<script setup lang="ts">
import type { ActivationRecord } from "./record-client";

import { ref } from "vue";

import { IconifyIcon } from "@vben/icons";

import { Image, ImagePreviewGroup, Space } from "ant-design-vue";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import { listActivations } from "./record-client";

const previewVisible = ref(false);
const previewImage = ref("");

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

function openPhotoPreview(row: ActivationRecord) {
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
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="激活记录">
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
