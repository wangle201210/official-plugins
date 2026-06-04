<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-record-checkin",
  title: "签到记录",
};
</script>

<script setup lang="ts">
import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import { listCheckins } from "./record-client";

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
      { field: "nickname", title: "玩家", minWidth: 160 },
      { field: "checkinDate", title: "签到日期", width: 140 },
      { field: "amount", title: "获得草量", width: 110 },
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
          return await listCheckins({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: { keyField: "id" },
    id: "sicau-niu-record-checkin-grid",
  },
});
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="签到记录" />
  </Page>
</template>
