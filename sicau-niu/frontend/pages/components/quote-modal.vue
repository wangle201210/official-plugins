<script setup lang="ts">
import type { QuoteItem } from "../quote-client";

import { computed, ref } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { message } from "ant-design-vue";

import { useVbenForm, z } from "#/adapter/form";

import { createQuote, updateQuote } from "../quote-client";

const emit = defineEmits<{ reload: [] }>();

interface QuoteFormValues {
  content: string;
  enabled: number;
}

const recordId = ref(0);
const formValues: QuoteFormValues = {
  content: "",
  enabled: 1,
};

const isEdit = computed(() => recordId.value > 0);
const title = computed(() => (isEdit.value ? "编辑金句" : "新增金句"));

const [QuoteForm, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: "w-full",
    },
    labelWidth: 80,
  },
  schema: [
    {
      component: "Textarea",
      componentProps: {
        "data-testid": "sicau-niu-quote-content-input",
        maxlength: 512,
        placeholder: "请输入金句内容",
        rows: 4,
      },
      fieldName: "content",
      label: "金句内容",
      rules: z.string().min(1, { message: "请输入金句内容" }),
    },
    {
      component: "Select",
      componentProps: {
        "data-testid": "sicau-niu-quote-enabled-select",
        options: [
          { label: "启用", value: 1 },
          { label: "停用", value: 0 },
        ],
      },
      defaultValue: 1,
      fieldName: "enabled",
      label: "是否启用",
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
    const values = await formApi.getValues<QuoteFormValues>();
    const payload = {
      content: values.content.trim(),
      enabled: values.enabled ?? 1,
    };
    if (isEdit.value) {
      await updateQuote(recordId.value, payload);
      message.success("更新成功");
    } else {
      await createQuote(payload);
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
  const data = modalApi.getData<Partial<QuoteItem>>();
  recordId.value = data?.id ?? 0;
  await formApi.setValues({
    content: data?.content ?? formValues.content,
    enabled: data?.enabled ?? formValues.enabled,
  });
}

async function handleClosed() {
  recordId.value = 0;
  await formApi.resetForm();
}
</script>

<template>
  <Modal :title="title">
    <div data-testid="sicau-niu-quote-form">
      <QuoteForm />
    </div>
  </Modal>
</template>
