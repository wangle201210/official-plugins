<script setup lang="ts">
import { ref, watch } from "vue";

import type { UploadFile } from "ant-design-vue/es/upload/interface";
import { Button as AButton, Upload as AUpload, message } from "ant-design-vue";

import { uploadApi } from "#/api";
import { $t } from "#/locales";

const props = withDefaults(
  defineProps<{
    disabled?: boolean;
    maxCount?: number;
    scene?: string;
    value?: string;
  }>(),
  {
    disabled: false,
    maxCount: 1,
    scene: "other",
    value: "",
  },
);

const emit = defineEmits<{
  "update:value": [value: string];
}>();

const fileList = ref<UploadFile[]>([]);
const uploading = ref(false);

watch(
  () => props.value,
  (value) => {
    fileList.value = value ? [uploadFileFromUrl(value)] : [];
  },
  { immediate: true },
);

function uploadFileFromUrl(url: string): UploadFile {
  const name = decodeURIComponent(url.split("/").pop() || url);
  return {
    fileName: name,
    name,
    status: "done",
    thumbUrl: url,
    uid: url,
    url,
  };
}

async function handleBeforeUpload(file: File) {
  uploading.value = true;
  try {
    const result = await uploadApi(file, { scene: props.scene });
    emit("update:value", result.url);
    fileList.value = [uploadFileFromUrl(result.url)];
  } catch {
    message.error($t("plugin.cms.editor.messages.imageUploadFailed"));
  } finally {
    uploading.value = false;
  }
  return false;
}

function handleRemove() {
  emit("update:value", "");
  fileList.value = [];
  return true;
}
</script>

<template>
  <AUpload
    accept="image/*"
    :before-upload="handleBeforeUpload"
    :disabled="disabled || uploading"
    :file-list="fileList"
    list-type="picture-card"
    :max-count="maxCount"
    @remove="handleRemove"
  >
    <AButton v-if="fileList.length < maxCount" :loading="uploading">
      {{ $t("pages.upload.button") }}
    </AButton>
  </AUpload>
</template>
