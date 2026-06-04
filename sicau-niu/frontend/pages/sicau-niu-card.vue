<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-card",
  title: "卡片管理",
};
</script>

<script setup lang="ts">
import type { CardItem } from "./card-client";

import { Popconfirm, Space } from "ant-design-vue";

import { useAccess } from "@vben/access";
import { useVbenModal } from "@vben/common-ui";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import CardModal from "./components/card-modal.vue";
import { deleteCard, listCards } from "./card-client";

const { hasAccessByCodes } = useAccess();

const pluginAccessCodes = {
  create: "sicau-niu:card:create",
  update: "sicau-niu:card:update",
  delete: "sicau-niu:card:delete",
} as const;

const categoryMap: Record<string, string> = {
  person: "人物",
  event: "事件",
  research: "科研",
  college: "院系",
  spirit: "精神",
};

function formatCategory(code: string) {
  return categoryMap[code] ?? "-";
}

const [RecordModal, recordModalApi] = useVbenModal({
  connectedComponent: CardModal,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "标题",
      },
      {
        component: "Select",
        fieldName: "category",
        label: "分类",
        componentProps: {
          options: [
            { label: "人物", value: "person" },
            { label: "事件", value: "event" },
            { label: "科研", value: "research" },
            { label: "院系", value: "college" },
            { label: "精神", value: "spirit" },
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
        field: "title",
        minWidth: 180,
        title: "标题",
      },
      {
        field: "category",
        formatter: ({ cellValue }) => formatCategory(cellValue),
        title: "分类",
        width: 90,
      },
      {
        field: "niuCode",
        minWidth: 120,
        title: "所属牛序号",
      },
      {
        field: "niuName",
        minWidth: 140,
        title: "所属牛名称",
      },
      {
        field: "imagePath",
        formatter: ({ cellValue }) => (cellValue ? "是" : "否"),
        title: "是否有图",
        width: 90,
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
          return await listCards({
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
    id: "sicau-niu-card-grid",
  },
});

function canCreateCard() {
  return hasAccessByCodes([pluginAccessCodes.create]);
}

function canUpdateCard() {
  return hasAccessByCodes([pluginAccessCodes.update]);
}

function canDeleteCard() {
  return hasAccessByCodes([pluginAccessCodes.delete]);
}

function handleAddCard() {
  recordModalApi.setData({});
  recordModalApi.open();
}

function handleEditCard(row: CardItem) {
  recordModalApi.setData({ id: row.id });
  recordModalApi.open();
}

async function handleDeleteCard(row: CardItem) {
  await deleteCard(row.id);
  await gridApi.query();
}

function handleReload() {
  gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="卡片管理">
      <template #toolbar-tools>
        <Space>
          <a-button
            v-if="canCreateCard()"
            data-testid="sicau-niu-card-add"
            type="primary"
            @click="handleAddCard"
          >
            新增
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button
            v-if="canUpdateCard()"
            :data-testid="`sicau-niu-card-edit-${row.id}`"
            @click.stop="handleEditCard(row)"
          >
            编辑
          </ghost-button>
          <Popconfirm
            v-if="canDeleteCard()"
            title="确定删除该卡片吗？"
            @confirm="handleDeleteCard(row)"
          >
            <ghost-button
              danger
              :data-testid="`sicau-niu-card-delete-${row.id}`"
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
