<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-record-gift",
  title: "送草记录",
};
</script>

<script setup lang="ts">
import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import { listGifts } from "./record-client";

const [Grid] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "InputNumber",
        fieldName: "fromUserId",
        label: "赠送人ID",
        componentProps: { min: 1, class: "w-full" },
      },
      {
        component: "InputNumber",
        fieldName: "toUserId",
        label: "接收人ID",
        componentProps: { min: 1, class: "w-full" },
      },
    ],
    commonConfig: { labelWidth: 80, componentProps: { allowClear: true } },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      { field: "fromNickname", title: "赠送人", minWidth: 140 },
      { field: "toNickname", title: "接收人", minWidth: 140 },
      { field: "amount", title: "草量", width: 90 },
      { field: "giftDate", title: "日期", width: 130 },
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
          return await listGifts({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: { keyField: "id" },
    id: "sicau-niu-record-gift-grid",
  },
});
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="送草记录" />
  </Page>
</template>
