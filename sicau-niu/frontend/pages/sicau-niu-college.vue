<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-college",
  title: "院系字典",
};
</script>

<script setup lang="ts">
import type { CollegeItem } from "./college-client";

import { Popconfirm, Space } from "ant-design-vue";

import { useAccess } from "@vben/access";
import { useVbenModal } from "@vben/common-ui";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import CollegeModal from "./components/college-modal.vue";
import { deleteCollege, listColleges } from "./college-client";

const { hasAccessByCodes } = useAccess();

const pluginAccessCodes = {
  create: "sicau-niu:college:create",
  update: "sicau-niu:college:update",
  delete: "sicau-niu:college:delete",
} as const;

const [RecordModal, recordModalApi] = useVbenModal({
  connectedComponent: CollegeModal,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        componentProps: { "data-testid": "sicau-niu-college-keyword-input" },
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
      {
        field: "name",
        minWidth: 200,
        title: "名称",
      },
      {
        field: "sort",
        title: "排序",
        width: 120,
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
          return await listColleges({
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
    id: "sicau-niu-college-grid",
  },
});

function canCreateCollege() {
  return hasAccessByCodes([pluginAccessCodes.create]);
}

function canUpdateCollege() {
  return hasAccessByCodes([pluginAccessCodes.update]);
}

function canDeleteCollege() {
  return hasAccessByCodes([pluginAccessCodes.delete]);
}

function handleAddCollege() {
  recordModalApi.setData({});
  recordModalApi.open();
}

function handleEditCollege(row: CollegeItem) {
  recordModalApi.setData({ id: row.id, name: row.name, sort: row.sort });
  recordModalApi.open();
}

async function handleDeleteCollege(row: CollegeItem) {
  await deleteCollege(row.id);
  await gridApi.query();
}

function handleReload() {
  gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="院系字典">
      <template #toolbar-tools>
        <Space>
          <a-button
            v-if="canCreateCollege()"
            data-testid="sicau-niu-college-add"
            type="primary"
            @click="handleAddCollege"
          >
            新增
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button
            v-if="canUpdateCollege()"
            :data-testid="`sicau-niu-college-edit-${row.id}`"
            @click.stop="handleEditCollege(row)"
          >
            编辑
          </ghost-button>
          <Popconfirm
            v-if="canDeleteCollege()"
            title="确定删除该院系吗？"
            @confirm="handleDeleteCollege(row)"
          >
            <ghost-button
              danger
              :data-testid="`sicau-niu-college-delete-${row.id}`"
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
