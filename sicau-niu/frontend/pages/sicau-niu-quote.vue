<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-quote",
  title: "金句管理",
};
</script>

<script setup lang="ts">
import type { QuoteItem } from "./quote-client";

import { Popconfirm, Space } from "ant-design-vue";

import { useAccess } from "@vben/access";
import { useVbenModal } from "@vben/common-ui";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import QuoteModal from "./components/quote-modal.vue";
import { deleteQuote, listQuotes } from "./quote-client";

const { hasAccessByCodes } = useAccess();

const pluginAccessCodes = {
  create: "sicau-niu:quote:create",
  update: "sicau-niu:quote:update",
  delete: "sicau-niu:quote:delete",
} as const;

const [RecordModal, recordModalApi] = useVbenModal({
  connectedComponent: QuoteModal,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "金句内容",
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
        field: "content",
        minWidth: 320,
        title: "金句内容",
      },
      {
        field: "enabled",
        formatter: ({ cellValue }) => (cellValue === 1 ? "启用" : "停用"),
        title: "是否启用",
        width: 100,
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
          return await listQuotes({
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
    id: "sicau-niu-quote-grid",
  },
});

function canCreateQuote() {
  return hasAccessByCodes([pluginAccessCodes.create]);
}

function canUpdateQuote() {
  return hasAccessByCodes([pluginAccessCodes.update]);
}

function canDeleteQuote() {
  return hasAccessByCodes([pluginAccessCodes.delete]);
}

function handleAddQuote() {
  recordModalApi.setData({});
  recordModalApi.open();
}

function handleEditQuote(row: QuoteItem) {
  recordModalApi.setData({
    id: row.id,
    content: row.content,
    enabled: row.enabled,
  });
  recordModalApi.open();
}

async function handleDeleteQuote(row: QuoteItem) {
  await deleteQuote(row.id);
  await gridApi.query();
}

function handleReload() {
  gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="金句管理">
      <template #toolbar-tools>
        <Space>
          <a-button
            v-if="canCreateQuote()"
            data-testid="sicau-niu-quote-add"
            type="primary"
            @click="handleAddQuote"
          >
            新增
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button
            v-if="canUpdateQuote()"
            :data-testid="`sicau-niu-quote-edit-${row.id}`"
            @click.stop="handleEditQuote(row)"
          >
            编辑
          </ghost-button>
          <Popconfirm
            v-if="canDeleteQuote()"
            title="确定删除该金句吗？"
            @confirm="handleDeleteQuote(row)"
          >
            <ghost-button
              danger
              :data-testid="`sicau-niu-quote-delete-${row.id}`"
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
