<script setup lang="ts">
import type { UploadProps } from "ant-design-vue";

import type { CardItem } from "../card-client";

import { computed, ref } from "vue";

import { useVbenModal } from "@vben/common-ui";
import { IconifyIcon } from "@vben/icons";

import { Image, message, Upload } from "ant-design-vue";

import { uploadApi } from "#/api/core";
import { useVbenForm, z } from "#/adapter/form";

import { createCard, getCard, updateCard } from "../card-client";

const emit = defineEmits<{ reload: [] }>();

interface CardFormValues {
  niuId: number;
  category: string;
  title: string;
  content: string;
}

const recordId = ref(0);
const imagePath = ref("");
const uploading = ref(false);

const isEdit = computed(() => recordId.value > 0);
const title = computed(() => (isEdit.value ? "编辑卡片" : "新增卡片"));

const [CardForm, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: "w-full",
    },
    labelWidth: 80,
  },
  schema: [
    {
      component: "InputNumber",
      componentProps: {
        "data-testid": "sicau-niu-card-niuid-input",
        class: "w-full",
        min: 1,
        placeholder: "请输入所属牛 ID",
        precision: 0,
      },
      fieldName: "niuId",
      label: "所属牛",
      rules: z.number({ message: "请输入所属牛 ID" }).min(1, {
        message: "请输入所属牛 ID",
      }),
    },
    {
      component: "Select",
      componentProps: {
        "data-testid": "sicau-niu-card-category-select",
        options: [
          { label: "人物", value: "person" },
          { label: "事件", value: "event" },
          { label: "科研", value: "research" },
          { label: "院系", value: "college" },
          { label: "精神", value: "spirit" },
        ],
        placeholder: "请选择分类",
      },
      fieldName: "category",
      label: "分类",
      rules: z.string().min(1, { message: "请选择卡片分类" }),
    },
    {
      component: "Input",
      componentProps: {
        "data-testid": "sicau-niu-card-title-input",
        maxlength: 128,
        placeholder: "请输入卡片标题",
      },
      fieldName: "title",
      label: "标题",
      rules: z.string().min(1, { message: "请输入卡片标题" }),
    },
    {
      component: "Textarea",
      componentProps: {
        "data-testid": "sicau-niu-card-content-input",
        maxlength: 1000,
        placeholder: "请输入卡片文案",
        rows: 4,
      },
      fieldName: "content",
      label: "文案",
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

const customUpload: UploadProps["customRequest"] = async (options) => {
  const { file, onError, onSuccess } = options;
  try {
    uploading.value = true;
    const result = await uploadApi(file as File, { scene: "other" });
    imagePath.value = result.url;
    message.success("图片上传成功");
    onSuccess?.(result);
  } catch (error) {
    message.error("图片上传失败");
    onError?.(error as any);
  } finally {
    uploading.value = false;
  }
};

function handleRemoveImage() {
  imagePath.value = "";
}

async function handleConfirm() {
  try {
    modalApi.lock(true);
    const { valid } = await formApi.validate();
    if (!valid) {
      return;
    }
    const values = await formApi.getValues<CardFormValues>();
    const payload = {
      niuId: values.niuId,
      category: values.category,
      title: values.title.trim(),
      content: values.content?.trim() ?? "",
      imagePath: imagePath.value,
    };
    if (isEdit.value) {
      await updateCard(recordId.value, payload);
      message.success("更新成功");
    } else {
      await createCard(payload);
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
  const data = modalApi.getData<Partial<CardItem>>();
  recordId.value = data?.id ?? 0;
  if (recordId.value > 0) {
    const detail = await getCard(recordId.value);
    imagePath.value = detail.imagePath;
    await formApi.setValues({
      niuId: detail.niuId,
      category: detail.category,
      title: detail.title,
      content: detail.content,
    });
  }
}

async function handleClosed() {
  recordId.value = 0;
  imagePath.value = "";
  await formApi.resetForm();
}
</script>

<template>
  <Modal :title="title">
    <div data-testid="sicau-niu-card-form">
      <CardForm />
      <div class="mt-2 px-2">
        <div class="mb-1 text-sm">卡片图片</div>
        <Upload.Dragger
          :custom-request="customUpload"
          :show-upload-list="false"
          accept="image/*"
          data-testid="sicau-niu-card-image-upload"
        >
          <p class="ant-upload-drag-icon flex justify-center">
            <IconifyIcon icon="ant-design:inbox-outlined" class="text-2xl" />
          </p>
          <p class="ant-upload-text">点击或拖拽图片到此区域上传</p>
          <p class="ant-upload-hint">仅支持单张图片，上传后保存为存储路径</p>
        </Upload.Dragger>
        <div v-if="imagePath" class="mt-2">
          <Image
            :src="imagePath"
            :width="120"
            :height="120"
            class="rounded border"
            :style="{ objectFit: 'cover' }"
            data-testid="sicau-niu-card-image-preview"
          />
          <div class="mt-1 flex items-center justify-between text-sm">
            <span class="truncate" data-testid="sicau-niu-card-image-path">
              {{ imagePath }}
            </span>
            <a-button
              type="link"
              size="small"
              danger
              :loading="uploading"
              @click="handleRemoveImage"
            >
              移除
            </a-button>
          </div>
        </div>
      </div>
    </div>
  </Modal>
</template>
