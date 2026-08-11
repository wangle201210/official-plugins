<script setup lang="ts">
import type { IronItem } from "../iron-client";

import { computed, ref } from "vue";

import { useVbenModal } from "@vben/common-ui";
import { IconifyIcon } from "@vben/icons";

import { Descriptions, DescriptionsItem, message } from "ant-design-vue";

import { useVbenForm, z } from "#/adapter/form";
import { formatTimestamp } from "#/utils/time";

import {
  createIron,
  setIronReportingCycleToTenSeconds,
  updateIron,
} from "../iron-client";

const emit = defineEmits<{ reload: [] }>();

interface IronFormValues {
  code: string;
  name: string;
  remark: string;
}

const recordId = ref(0);
const reportingCycleUpdating = ref(false);
const snapshot = ref<Pick<IronItem, "lastLat" | "lastLng" | "locatedAt">>({
  lastLat: 0,
  lastLng: 0,
  locatedAt: null,
});
const formValues: IronFormValues = {
  code: "",
  name: "",
  remark: "",
};

const isEdit = computed(() => recordId.value > 0);
const title = computed(() => (isEdit.value ? "编辑铁牛" : "新增铁牛"));

const [IronForm, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: "w-full",
    },
    labelWidth: 80,
  },
  schema: [
    {
      component: "Input",
      componentProps: {
        "data-testid": "sicau-niu-iron-code-input",
        maxlength: 64,
        placeholder: "请输入铁牛标识",
      },
      fieldName: "code",
      label: "标识",
      rules: z.string().min(1, { message: "请输入铁牛标识" }),
    },
    {
      component: "Input",
      componentProps: {
        "data-testid": "sicau-niu-iron-name-input",
        maxlength: 128,
        placeholder: "请输入铁牛名称",
      },
      fieldName: "name",
      label: "名称",
      rules: z.string().min(1, { message: "请输入铁牛名称" }),
    },
    {
      component: "Textarea",
      componentProps: {
        "data-testid": "sicau-niu-iron-remark-input",
        maxlength: 255,
        placeholder: "请输入备注",
        rows: 3,
      },
      fieldName: "remark",
      label: "备注",
    },
  ],
  showDefaultActions: false,
});

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  onClosed: handleClosed,
  onConfirm: handleConfirm,
  onOpenChange: handleOpenChange,
});

async function handleConfirm() {
  try {
    modalApi.lock(true);
    const { valid } = await formApi.validate();
    if (!valid) {
      return;
    }
    const values = await formApi.getValues<IronFormValues>();
    const payload = {
      code: values.code.trim(),
      name: values.name.trim(),
      remark: values.remark?.trim() ?? "",
    };
    if (isEdit.value) {
      await updateIron(recordId.value, payload);
      message.success("更新成功");
    } else {
      await createIron(payload);
      message.success("新增成功");
    }
    emit("reload");
    modalApi.close();
  } finally {
    modalApi.lock(false);
  }
}

async function handleUpdateReportingCycle() {
  const { valid } = await formApi.validateField("code");
  if (!valid) {
    return;
  }
  const { code } = await formApi.getValues<Pick<IronFormValues, "code">>();

  try {
    reportingCycleUpdating.value = true;
    modalApi.lock(true);
    await setIronReportingCycleToTenSeconds(code.trim());
    message.success("上报频率已更新为 10 秒");
  } finally {
    reportingCycleUpdating.value = false;
    modalApi.lock(false);
  }
}

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<Partial<IronItem>>();
  recordId.value = data?.id ?? 0;
  snapshot.value = {
    lastLat: data?.lastLat ?? 0,
    lastLng: data?.lastLng ?? 0,
    locatedAt: data?.locatedAt ?? null,
  };
  await formApi.setValues({
    code: data?.code ?? formValues.code,
    name: data?.name ?? formValues.name,
    remark: data?.remark ?? formValues.remark,
  });
}

async function handleClosed() {
  recordId.value = 0;
  snapshot.value = {
    lastLat: 0,
    lastLng: 0,
    locatedAt: null,
  };
  await formApi.resetForm();
}

function formatCoordinate(value: number) {
  if (!value || Number.isNaN(value)) {
    return "-";
  }
  return value.toFixed(6);
}

function formatLocatedAt(value: number | null) {
  return value ? formatTimestamp(value) : "-";
}
</script>

<template>
  <Modal :title="title">
    <div data-testid="sicau-niu-iron-form">
      <IronForm />
      <div v-if="!isEdit" class="mt-3 flex justify-end">
        <a-button
          class="w-[190px]"
          data-testid="sicau-niu-iron-reporting-cycle"
          :loading="reportingCycleUpdating"
          @click="handleUpdateReportingCycle"
        >
          <IconifyIcon class="mr-1" icon="ant-design:sync-outlined" />
          设置 10 秒上报频率
        </a-button>
      </div>
      <Descriptions
        v-if="isEdit"
        bordered
        class="mt-4"
        data-testid="sicau-niu-iron-location-snapshot"
        size="small"
        title="定位信息"
        :column="1"
      >
        <DescriptionsItem label="纬度">
          {{ formatCoordinate(snapshot.lastLat) }}
        </DescriptionsItem>
        <DescriptionsItem label="经度">
          {{ formatCoordinate(snapshot.lastLng) }}
        </DescriptionsItem>
        <DescriptionsItem label="最近同步时间">
          {{ formatLocatedAt(snapshot.locatedAt) }}
        </DescriptionsItem>
      </Descriptions>
    </div>
  </Modal>
</template>
