<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-honor",
  title: "荣誉配置",
};
</script>

<script setup lang="ts">
import type { HonorItem } from "./honor-client";

import { Popconfirm, Space } from "ant-design-vue";

import { useAccess } from "@vben/access";
import { useVbenModal } from "@vben/common-ui";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { Page } from "#/plugins/dynamic";

import HonorModal from "./components/honor-modal.vue";
import { deleteHonor, listHonors } from "./honor-client";

const { hasAccessByCodes } = useAccess();

const pluginAccessCodes = {
  create: "sicau-niu:honor:create",
  update: "sicau-niu:honor:update",
  delete: "sicau-niu:honor:delete",
} as const;

const honorTypeLabels: Record<string, string> = {
  badge: "徽章",
  avatar_frame: "头像框",
  certificate: "证书",
};

const unlockTypeLabels: Record<string, string> = {
  participation: "参与即得",
  feed_count: "喂草次数",
  activation_count: "激活数",
  category_complete: "集齐分类",
  full_complete: "集齐全套",
};

const [RecordModal, recordModalApi] = useVbenModal({
  connectedComponent: HonorModal,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "编码/名称",
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
        minWidth: 160,
        title: "编码",
      },
      {
        field: "name",
        minWidth: 160,
        title: "名称",
      },
      {
        field: "honorType",
        formatter: ({ cellValue }) => honorTypeLabels[cellValue] ?? cellValue,
        title: "类型",
        width: 120,
      },
      {
        field: "unlockType",
        formatter: ({ cellValue }) => unlockTypeLabels[cellValue] ?? cellValue,
        title: "解锁规则",
        width: 120,
      },
      {
        field: "threshold",
        title: "阈值",
        width: 100,
      },
      {
        field: "sort",
        title: "排序",
        width: 90,
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
          return await listHonors({
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
    id: "sicau-niu-honor-grid",
  },
});

function canCreateHonor() {
  return hasAccessByCodes([pluginAccessCodes.create]);
}

function canUpdateHonor() {
  return hasAccessByCodes([pluginAccessCodes.update]);
}

function canDeleteHonor() {
  return hasAccessByCodes([pluginAccessCodes.delete]);
}

function handleAddHonor() {
  recordModalApi.setData({});
  recordModalApi.open();
}

function handleEditHonor(row: HonorItem) {
  recordModalApi.setData({ id: row.id });
  recordModalApi.open();
}

async function handleDeleteHonor(row: HonorItem) {
  await deleteHonor(row.id);
  await gridApi.query();
}

function handleReload() {
  gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="荣誉配置">
      <template #toolbar-tools>
        <Space>
          <a-button
            v-if="canCreateHonor()"
            data-testid="sicau-niu-honor-add"
            type="primary"
            @click="handleAddHonor"
          >
            新增
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button
            v-if="canUpdateHonor()"
            :data-testid="`sicau-niu-honor-edit-${row.id}`"
            @click.stop="handleEditHonor(row)"
          >
            编辑
          </ghost-button>
          <Popconfirm
            v-if="canDeleteHonor()"
            title="确定删除该荣誉吗？"
            @confirm="handleDeleteHonor(row)"
          >
            <ghost-button
              danger
              :data-testid="`sicau-niu-honor-delete-${row.id}`"
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
