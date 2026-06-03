<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-sidebar-entry",
  title: "sicau-niu 示例",
};
</script>

<script setup lang="ts">
import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { Page } from "#/plugins/dynamic";
import { formatTimestamp } from "#/utils/time";

import { listCattle } from "./niu-client";

const [Grid] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "名称",
      },
    ],
    commonConfig: {
      labelWidth: 80,
      componentProps: {
        allowClear: true,
      },
    },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      { field: "name", minWidth: 160, title: "名称" },
      { field: "breed", minWidth: 160, title: "品种" },
      { field: "weightKg", width: 120, title: "体重(kg)" },
      {
        field: "createdAt",
        formatter: ({ cellValue }) => formatTimestamp(cellValue),
        title: "创建时间",
        width: 200,
      },
    ],
    height: "auto",
    keepSource: true,
    pagerConfig: { enabled: false },
    proxyConfig: {
      ajax: {
        query: async (_params: unknown, formValues = {}) => {
          return await listCattle({ ...formValues });
        },
      },
    },
    rowConfig: {
      keyField: "id",
    },
    id: "sicau-niu-cattle-grid",
  },
});
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="sicau-niu 示例牛只列表" />
  </Page>
</template>
