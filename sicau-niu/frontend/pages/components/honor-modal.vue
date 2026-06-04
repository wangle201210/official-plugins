<script setup lang="ts">
import type { HonorItem } from "../honor-client";

import { computed, ref } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { message } from "ant-design-vue";

import { useVbenForm, z } from "#/adapter/form";

import { createHonor, getHonor, updateHonor } from "../honor-client";

const emit = defineEmits<{ reload: [] }>();

interface HonorFormValues {
  honorType: string;
  code: string;
  name: string;
  unlockType: string;
  threshold: number | null;
  category: string;
  imagePath: string;
  sort: number | null;
}

const recordId = ref(0);

const isEdit = computed(() => recordId.value > 0);
const title = computed(() => (isEdit.value ? "编辑荣誉" : "新增荣誉"));

const [HonorForm, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: "w-full",
    },
    labelWidth: 90,
  },
  schema: [
    {
      component: "Select",
      componentProps: {
        "data-testid": "sicau-niu-honor-type-select",
        options: [
          { label: "徽章", value: "badge" },
          { label: "头像框", value: "avatar_frame" },
          { label: "证书", value: "certificate" },
        ],
      },
      defaultValue: "badge",
      fieldName: "honorType",
      label: "类型",
      rules: z.string().min(1, { message: "请选择荣誉类型" }),
    },
    {
      component: "Input",
      componentProps: {
        "data-testid": "sicau-niu-honor-code-input",
        maxlength: 64,
        placeholder: "请输入荣誉编码",
      },
      fieldName: "code",
      label: "编码",
      rules: z.string().min(1, { message: "请输入荣誉编码" }),
    },
    {
      component: "Input",
      componentProps: {
        "data-testid": "sicau-niu-honor-name-input",
        maxlength: 64,
        placeholder: "请输入荣誉名称",
      },
      fieldName: "name",
      label: "名称",
      rules: z.string().min(1, { message: "请输入荣誉名称" }),
    },
    {
      component: "Select",
      componentProps: {
        "data-testid": "sicau-niu-honor-unlock-select",
        options: [
          { label: "参与即得", value: "participation" },
          { label: "喂草次数", value: "feed_count" },
          { label: "激活数", value: "activation_count" },
          { label: "集齐分类", value: "category_complete" },
          { label: "集齐全套", value: "full_complete" },
        ],
      },
      defaultValue: "participation",
      fieldName: "unlockType",
      label: "解锁规则",
      rules: z.string().min(1, { message: "请选择解锁规则" }),
    },
    {
      component: "InputNumber",
      componentProps: {
        "data-testid": "sicau-niu-honor-threshold-input",
        class: "w-full",
        min: 0,
        placeholder: "计数类规则的阈值",
        precision: 0,
      },
      defaultValue: 0,
      fieldName: "threshold",
      label: "阈值",
    },
    {
      component: "Select",
      componentProps: {
        "data-testid": "sicau-niu-honor-category-select",
        allowClear: true,
        options: [
          { label: "人物", value: "person" },
          { label: "事件", value: "event" },
          { label: "科研", value: "research" },
          { label: "学院", value: "college" },
          { label: "精神", value: "spirit" },
        ],
        placeholder: "集齐分类规则的卡片分类",
      },
      dependencies: {
        show: (values) => values.unlockType === "category_complete",
        triggerFields: ["unlockType"],
      },
      fieldName: "category",
      label: "分类",
    },
    {
      component: "Input",
      componentProps: {
        "data-testid": "sicau-niu-honor-image-input",
        maxlength: 255,
        placeholder: "可选，宿主 /upload 路径或留空",
      },
      fieldName: "imagePath",
      label: "图片路径",
    },
    {
      component: "InputNumber",
      componentProps: {
        "data-testid": "sicau-niu-honor-sort-input",
        class: "w-full",
        precision: 0,
      },
      defaultValue: 0,
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
    const values = await formApi.getValues<HonorFormValues>();
    const isCategoryComplete = values.unlockType === "category_complete";
    const payload = {
      honorType: values.honorType,
      code: values.code.trim(),
      name: values.name.trim(),
      unlockType: values.unlockType,
      threshold: values.threshold ?? 0,
      category: isCategoryComplete ? (values.category ?? "") : "",
      imagePath: values.imagePath?.trim() ?? "",
      sort: values.sort ?? 0,
    };
    if (isEdit.value) {
      await updateHonor(recordId.value, payload);
      message.success("更新成功");
    } else {
      await createHonor(payload);
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
  const data = modalApi.getData<Partial<HonorItem>>();
  recordId.value = data?.id ?? 0;
  if (recordId.value > 0) {
    const detail = await getHonor(recordId.value);
    await formApi.setValues({
      honorType: detail.honorType,
      code: detail.code,
      name: detail.name,
      unlockType: detail.unlockType,
      threshold: detail.threshold,
      category: detail.category,
      imagePath: detail.imagePath,
      sort: detail.sort,
    });
  }
}

async function handleClosed() {
  recordId.value = 0;
  await formApi.resetForm();
}
</script>

<template>
  <Modal :title="title">
    <div data-testid="sicau-niu-honor-form">
      <HonorForm />
    </div>
  </Modal>
</template>
