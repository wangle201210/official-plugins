<script setup lang="ts">
import type { CollegeItem } from "../college-client";

import { computed, ref } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { message } from "ant-design-vue";

import { useVbenForm, z } from "#/adapter/form";

import { createCollege, updateCollege } from "../college-client";

const emit = defineEmits<{ reload: [] }>();

interface CollegeFormValues {
  name: string;
  sort: number;
}

const recordId = ref(0);
const formValues: CollegeFormValues = {
  name: "",
  sort: 0,
};

const isEdit = computed(() => recordId.value > 0);
const title = computed(() => (isEdit.value ? "编辑院系" : "新增院系"));

const [CollegeForm, formApi] = useVbenForm({
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
        "data-testid": "sicau-niu-college-name-input",
        maxlength: 64,
        placeholder: "请输入院系名称",
      },
      fieldName: "name",
      label: "名称",
      rules: z.string().min(1, { message: "请输入院系名称" }),
    },
    {
      component: "InputNumber",
      componentProps: {
        "data-testid": "sicau-niu-college-sort-input",
        class: "w-full",
        min: 0,
        precision: 0,
      },
      fieldName: "sort",
      label: "排序",
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
    const values = await formApi.getValues<CollegeFormValues>();
    const payload = {
      name: values.name.trim(),
      sort: values.sort ?? 0,
    };
    if (isEdit.value) {
      await updateCollege(recordId.value, payload);
      message.success("更新成功");
    } else {
      await createCollege(payload);
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
  const data = modalApi.getData<Partial<CollegeItem>>();
  recordId.value = data?.id ?? 0;
  await formApi.setValues({
    name: data?.name ?? formValues.name,
    sort: data?.sort ?? formValues.sort,
  });
}

async function handleClosed() {
  recordId.value = 0;
  await formApi.resetForm();
}
</script>

<template>
  <Modal :title="title">
    <div data-testid="sicau-niu-college-form">
      <CollegeForm />
    </div>
  </Modal>
</template>
