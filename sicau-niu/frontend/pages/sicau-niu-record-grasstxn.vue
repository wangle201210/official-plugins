<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-record-grasstxn",
  title: "草账户流水",
};
</script>

<script setup lang="ts">
import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import { listGrassTxns } from "./record-client";

const txnTypeMap: Record<string, string> = {
  checkin: "签到",
  feed: "喂草",
  steal: "偷草",
  gift: "送草",
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
    ],
    commonConfig: { labelWidth: 70, componentProps: { allowClear: true } },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      { field: "nickname", title: "玩家", minWidth: 150 },
      {
        field: "txnType",
        title: "类型",
        width: 100,
        formatter: ({ cellValue }) => txnTypeMap[cellValue] ?? cellValue,
      },
      { field: "delta", title: "增减量", width: 100 },
      { field: "refId", title: "关联ID", width: 100 },
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
          return await listGrassTxns({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: { keyField: "id" },
    id: "sicau-niu-record-grasstxn-grid",
  },
});
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="草账户流水" />
  </Page>
</template>
