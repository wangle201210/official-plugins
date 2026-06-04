<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-iron",
  title: "铁牛管理",
};
</script>

<script setup lang="ts">
import type { IronItem } from "./iron-client";

import { Popconfirm, Space } from "ant-design-vue";

import { useAccess } from "@vben/access";
import { useVbenModal } from "@vben/common-ui";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import IronModal from "./components/iron-modal.vue";
import { deleteIron, listIron } from "./iron-client";

const { hasAccessByCodes } = useAccess();

const pluginAccessCodes = {
  create: "sicau-niu:iron:create",
  update: "sicau-niu:iron:update",
  delete: "sicau-niu:iron:delete",
} as const;

const [RecordModal, recordModalApi] = useVbenModal({
  connectedComponent: IronModal,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "标识/名称",
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
        field: "code",
        minWidth: 140,
        title: "标识",
      },
      {
        field: "name",
        minWidth: 160,
        title: "名称",
      },
      {
        field: "locatedAt",
        formatter: ({ cellValue }) =>
          cellValue ? formatTimestamp(cellValue) : "-",
        title: "最近定位时间",
        width: 180,
      },
      {
        field: "remark",
        minWidth: 160,
        title: "备注",
      },
      {
        field: "createdAt",
        formatter: ({ cellValue }) => formatTimestamp(cellValue),
        title: "创建时间",
        width: 180,
      },
      {
        field: "action",
        fixed: "right",
        slots: { default: "action" },
        title: "操作",
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
          return await listIron({
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
    id: "sicau-niu-iron-grid",
  },
});

function canCreateIron() {
  return hasAccessByCodes([pluginAccessCodes.create]);
}

function canUpdateIron() {
  return hasAccessByCodes([pluginAccessCodes.update]);
}

function canDeleteIron() {
  return hasAccessByCodes([pluginAccessCodes.delete]);
}

function handleAddIron() {
  recordModalApi.setData({});
  recordModalApi.open();
}

function handleEditIron(row: IronItem) {
  recordModalApi.setData({
    id: row.id,
    code: row.code,
    name: row.name,
    remark: row.remark,
  });
  recordModalApi.open();
}

async function handleDeleteIron(row: IronItem) {
  await deleteIron(row.id);
  await gridApi.query();
}

function handleReload() {
  gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="铁牛管理">
      <template #toolbar-tools>
        <Space>
          <a-button
            v-if="canCreateIron()"
            data-testid="sicau-niu-iron-add"
            type="primary"
            @click="handleAddIron"
          >
            新增
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button
            v-if="canUpdateIron()"
            :data-testid="`sicau-niu-iron-edit-${row.id}`"
            @click.stop="handleEditIron(row)"
          >
            编辑
          </ghost-button>
          <Popconfirm
            v-if="canDeleteIron()"
            title="确定删除该铁牛吗？"
            @confirm="handleDeleteIron(row)"
          >
            <ghost-button
              danger
              :data-testid="`sicau-niu-iron-delete-${row.id}`"
              @click.stop=""
            >
              删除
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>
    <RecordModal @reload="handleReload" />
  </Page>
</template>
