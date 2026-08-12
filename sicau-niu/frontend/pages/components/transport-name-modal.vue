<script setup lang="ts">
import { ref } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { message } from "ant-design-vue";

import { useVbenForm, z } from "#/adapter/form";

import {
  transportRenameErrorMessage,
  updateTransportTeamName,
} from "../transport-client";

const emit = defineEmits<{ reload: [] }>();

const teamId = ref("");

const [TeamNameForm, formApi] = useVbenForm({
  commonConfig: {
    componentProps: { class: "w-full" },
    labelWidth: 72,
  },
  schema: [
    {
      component: "Input",
      componentProps: {
        "data-testid": "sicau-niu-transport-name-input",
        maxlength: 64,
        placeholder: "请输入未被有效团使用的名称",
      },
      fieldName: "name",
      label: "团名称",
      rules: z.string().trim().min(1, { message: "请输入团名称" }).max(64),
    },
  ],
  showDefaultActions: false,
});

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  onClosed: async () => {
    teamId.value = "";
    await formApi.resetForm();
  },
  onConfirm: async () => {
    try {
      modalApi.lock(true);
      const { valid } = await formApi.validate();
      if (!valid) {
        return;
      }
      const values = await formApi.getValues<{ name: string }>();
      await updateTransportTeamName(teamId.value, values.name.trim());
      message.success("团名称已更新");
      emit("reload");
      modalApi.close();
    } catch (error) {
      message.error(transportRenameErrorMessage(error));
    } finally {
      modalApi.lock(false);
    }
  },
  onOpenChange: async (open) => {
    if (!open) {
      return;
    }
    const data = modalApi.getData<{ id: string; name: string }>();
    teamId.value = data.id;
    await formApi.setValues({ name: data.name });
  },
});
</script>

<template>
  <Modal title="修改团名称">
    <TeamNameForm data-testid="sicau-niu-transport-name-form" />
  </Modal>
</template>
