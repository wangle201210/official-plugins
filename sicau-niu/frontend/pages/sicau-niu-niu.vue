<script lang="ts">
export const pluginPageMeta = {
  routePath: "sicau-niu-niu",
  title: "牛管理",
};
</script>

<script setup lang="ts">
import type { NiuItem } from "./niu-client";

import { Popconfirm, Space } from "ant-design-vue";

import { useAccess } from "@vben/access";
import { useVbenModal } from "@vben/common-ui";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { formatTimestamp } from "#/utils/time";
import { Page } from "#/plugins/dynamic";

import NiuModal from "./components/niu-modal.vue";
import { deleteNiu, listNiu } from "./niu-client";

const { hasAccessByCodes } = useAccess();

const pluginAccessCodes = {
  create: "sicau-niu:niu:create",
  update: "sicau-niu:niu:update",
  delete: "sicau-niu:niu:delete",
} as const;

const niuTypeMap: Record<string, string> = {
  common: "普通",
  special: "特殊",
};

const specialSubtypeMap: Record<string, string> = {
  college: "学院",
  contribution: "贡献",
  alumni: "校友",
  spirit: "精神",
};

const releaseStageMap: Record<string, string> = {
  warmup: "预热",
  main: "主体",
  climax: "高潮",
  closing: "收尾",
};

function formatNiuType(code: string) {
  return niuTypeMap[code] ?? "-";
}

function formatSpecialSubtype(code: string) {
  return specialSubtypeMap[code] ?? "-";
}

function formatReleaseStage(code: string) {
  return releaseStageMap[code] ?? "-";
}

const [RecordModal, recordModalApi] = useVbenModal({
  connectedComponent: NiuModal,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "序号/名称",
      },
      {
        component: "Select",
        fieldName: "niuType",
        label: "类型",
        componentProps: {
          options: [
            { label: "普通", value: "common" },
            { label: "特殊", value: "special" },
          ],
        },
      },
      {
        component: "Select",
        fieldName: "releaseStage",
        label: "上线阶段",
        componentProps: {
          options: [
            { label: "预热", value: "warmup" },
            { label: "主体", value: "main" },
            { label: "高潮", value: "climax" },
            { label: "收尾", value: "closing" },
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
        field: "code",
        minWidth: 120,
        title: "序号",
      },
      {
        field: "niuType",
        formatter: ({ cellValue }) => formatNiuType(cellValue),
        title: "类型",
        width: 90,
      },
      {
        field: "specialSubtype",
        formatter: ({ cellValue }) => formatSpecialSubtype(cellValue),
        title: "特殊子类",
        width: 100,
      },
      {
        field: "name",
        minWidth: 160,
        title: "名称",
      },
      {
        field: "collegeName",
        minWidth: 140,
        title: "所属院系",
      },
      {
        field: "hasCard",
        formatter: ({ cellValue }) => (cellValue ? "是" : "否"),
        title: "是否绑卡",
        width: 90,
      },
      {
        field: "releaseStage",
        formatter: ({ cellValue }) => formatReleaseStage(cellValue),
        title: "上线阶段",
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
          return await listNiu({
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
    id: "sicau-niu-niu-grid",
  },
});

function canCreateNiu() {
  return hasAccessByCodes([pluginAccessCodes.create]);
}

function canUpdateNiu() {
  return hasAccessByCodes([pluginAccessCodes.update]);
}

function canDeleteNiu() {
  return hasAccessByCodes([pluginAccessCodes.delete]);
}

function handleAddNiu() {
  recordModalApi.setData({});
  recordModalApi.open();
}

function handleEditNiu(row: NiuItem) {
  recordModalApi.setData({ id: row.id });
  recordModalApi.open();
}

async function handleDeleteNiu(row: NiuItem) {
  await deleteNiu(row.id);
  await gridApi.query();
}

function handleReload() {
  gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="牛管理">
      <template #toolbar-tools>
        <Space>
          <a-button
            v-if="canCreateNiu()"
            data-testid="sicau-niu-niu-add"
            type="primary"
            @click="handleAddNiu"
          >
            新增
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button
            v-if="canUpdateNiu()"
            :data-testid="`sicau-niu-niu-edit-${row.id}`"
            @click.stop="handleEditNiu(row)"
          >
            编辑
          </ghost-button>
          <Popconfirm
            v-if="canDeleteNiu()"
            title="确定删除该牛吗？将同时删除其主卡。"
            @confirm="handleDeleteNiu(row)"
          >
            <ghost-button
              danger
              :data-testid="`sicau-niu-niu-delete-${row.id}`"
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
