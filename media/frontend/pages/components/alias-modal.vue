<script setup lang="ts">
import type { MediaAliasInput } from "../media-client";

import { computed, reactive } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { Form, FormItem, Input, message, RadioGroup } from "ant-design-vue";

import {
  createMediaAlias,
  getMediaAlias,
  updateMediaAlias,
} from "../media-client";

const emit = defineEmits<{ reload: [] }>();

interface AliasFormData extends MediaAliasInput {
  id?: number;
}

const defaultValues: AliasFormData = {
  alias: "",
  autoRemove: 0,
  channelId: "",
  deviceId: "",
  streamPath: "",
};

const formData = reactive<AliasFormData>({ ...defaultValues });
const isEdit = computed(() => !!formData.id);
const modalTitle = computed(() => (isEdit.value ? "编辑流别名" : "新增流别名"));

const formRules = reactive({
  alias: [{ message: "请输入流别名", required: true }],
  channelId: [{ message: "请输入设备通道 ID", required: true }],
  deviceId: [{ message: "请输入设备 ID", required: true }],
  streamPath: [{ message: "请输入真实流路径", required: true }],
});

const { resetFields, validate, validateInfos } = Form.useForm(
  formData,
  formRules,
);

function setFormData(values: Partial<AliasFormData> = {}) {
  Object.assign(formData, defaultValues, values);
  if (values.id === undefined) {
    delete formData.id;
  }
}

const [Modal, modalApi] = useVbenModal({
  class: "w-[620px]",
  onConfirm: handleConfirm,
  onOpenChange: async (isOpen: boolean) => {
    if (!isOpen) return;
    const data = modalApi.getData();
    resetFields();
    setFormData();
    if (data?.id) {
      modalApi.setState({ confirmLoading: true });
      try {
        const record = await getMediaAlias(data.id);
        setFormData({
          id: record.id,
          alias: record.alias,
          autoRemove: record.autoRemove,
          channelId: record.channelId,
          deviceId: record.deviceId,
          streamPath: record.streamPath,
        });
      } finally {
        modalApi.setState({ confirmLoading: false });
      }
    }
  },
});

async function handleConfirm() {
  try {
    modalApi.lock(true);
    await validate();
    const { id, ...values } = {
      ...formData,
      alias: formData.alias.trim(),
      channelId: formData.channelId.trim(),
      deviceId: formData.deviceId.trim(),
      streamPath: formData.streamPath.trim(),
    };
    if (isEdit.value && id) {
      await updateMediaAlias(id, values);
      message.success("流别名已更新");
    } else {
      await createMediaAlias(values);
      message.success("流别名已创建");
    }
    emit("reload");
    modalApi.close();
  } finally {
    modalApi.lock(false);
  }
}
</script>

<template>
  <Modal :title="modalTitle">
    <Form layout="vertical">
      <FormItem label="流别名" v-bind="validateInfos.alias">
        <Input
          data-testid="media-alias-name"
          v-model:value="formData.alias"
          allow-clear
          placeholder="例如：camera-01"
        />
      </FormItem>
      <FormItem label="真实流路径" v-bind="validateInfos.streamPath">
        <Input
          data-testid="media-alias-stream-path"
          v-model:value="formData.streamPath"
          allow-clear
          placeholder="例如：live/camera-01"
        />
      </FormItem>
      <div class="grid gap-4 md:grid-cols-2">
        <FormItem label="设备 ID" v-bind="validateInfos.deviceId">
          <Input
            data-testid="media-alias-device-id"
            v-model:value="formData.deviceId"
            allow-clear
            :maxlength="64"
            placeholder="例如：34020000001320000001"
          />
        </FormItem>
        <FormItem label="设备通道 ID" v-bind="validateInfos.channelId">
          <Input
            data-testid="media-alias-channel-id"
            v-model:value="formData.channelId"
            allow-clear
            :maxlength="64"
            placeholder="例如：34020000001320000001"
          />
        </FormItem>
      </div>
      <FormItem label="自动移除">
        <RadioGroup
          data-testid="media-alias-auto-remove"
          v-model:value="formData.autoRemove"
          button-style="solid"
          option-type="button"
          :options="[
            { label: '是', value: 1 },
            { label: '否', value: 0 },
          ]"
        />
      </FormItem>
    </Form>
  </Modal>
</template>
