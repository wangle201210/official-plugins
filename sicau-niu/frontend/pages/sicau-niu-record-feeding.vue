<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-record-feeding",
  title: "喂草记录",
};
</script>

<script setup lang="ts">
import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import { listFeedings } from "./record-client";

const [Grid] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "InputNumber",
        fieldName: "niuId",
        label: "牛ID",
        componentProps: { min: 1, class: "w-full" },
      },
      {
        component: "InputNumber",
        fieldName: "userId",
        label: "玩家ID",
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
      { field: "baseAmount", title: "原始量", width: 90 },
      { field: "effectAmount", title: "实际效果", width: 100 },
      {
        field: "isIronBonus",
        title: "铁牛加成",
        width: 100,
        formatter: ({ cellValue }) => (cellValue === 1 ? "是" : "否"),
      },
      {
        field: "createdAt",
        title: "时间",
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
          return await listFeedings({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: { keyField: "id" },
    id: "sicau-niu-record-feeding-grid",
  },
});
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="喂草记录" />
  </Page>
</template>
