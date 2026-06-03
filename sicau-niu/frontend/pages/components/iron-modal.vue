<script setup lang="ts">
import type { IronItem } from "../iron-client";

import { computed, ref } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { message } from "ant-design-vue";

import { useVbenForm, z } from "#/adapter/form";

import { createIron, updateIron } from "../iron-client";

const emit = defineEmits<{ reload: [] }>();

interface IronFormValues {
  code: string;
  name: string;
  remark: string;
}

const recordId = ref(0);
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

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<Partial<IronItem>>();
  recordId.value = data?.id ?? 0;
  await formApi.setValues({
    code: data?.code ?? formValues.code,
    name: data?.name ?? formValues.name,
    remark: data?.remark ?? formValues.remark,
  });
}

async function handleClosed() {
  recordId.value = 0;
  await formApi.resetForm();
}
</script>

<template>
  <Modal :title="title">
    <div data-testid="sicau-niu-iron-form">
      <IronForm />
    </div>
  </Modal>
</template>
