<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-player",
  title: "寻牛玩家",
};
</script>

<script setup lang="ts">
import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import { listPlayers } from "./player-client";

const identityTypeMap: Record<string, string> = {
  student: "在校生",
  alumni: "校友",
  friend: "川农好友",
};

function formatIdentityType(code: string) {
  return identityTypeMap[code] ?? "-";
}

const [Grid] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "昵称",
      },
      {
        component: "Select",
        fieldName: "identityType",
        label: "身份",
        componentProps: {
          options: [
            { label: "在校生", value: "student" },
            { label: "校友", value: "alumni" },
            { label: "川农好友", value: "friend" },
          ],
        },
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
      {
        field: "nickname",
        minWidth: 160,
        title: "昵称",
      },
      {
        field: "phone",
        minWidth: 140,
        title: "手机号",
      },
      {
        field: "identityType",
        formatter: ({ cellValue }) => formatIdentityType(cellValue),
        title: "身份",
        width: 120,
      },
      {
        field: "grade",
        title: "年级",
        width: 100,
      },
      {
        field: "graduationYear",
        title: "毕业年",
        width: 100,
      },
      {
        field: "createdAt",
        formatter: ({ cellValue }) => formatTimestamp(cellValue),
        title: "创建时间",
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
          formValues = {},
        ) => {
          return await listPlayers({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: "id",
    },
    id: "sicau-niu-player-grid",
  },
});
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="寻牛玩家" />
  </Page>
</template>
