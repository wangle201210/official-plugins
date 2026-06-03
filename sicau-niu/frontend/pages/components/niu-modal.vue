<script setup lang="ts">
import type { NiuItem } from "../niu-client";

import { computed, ref } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { message } from "ant-design-vue";

import { useVbenForm, z } from "#/adapter/form";

import { listColleges } from "../college-client";
import { createNiu, getNiu, updateNiu } from "../niu-client";

const emit = defineEmits<{ reload: [] }>();

interface NiuFormValues {
  code: string;
  niuType: string;
  specialSubtype: string;
  name: string;
  collegeId: number | null;
  lat: number | null;
  lng: number | null;
  releaseStage: string;
  onlineAt: number | null;
  visibleWeekdays: string;
  visibleStart: string;
  visibleEnd: string;
}

const recordId = ref(0);

const isEdit = computed(() => recordId.value > 0);
const title = computed(() => (isEdit.value ? "编辑牛" : "新增牛"));

const [NiuForm, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: "w-full",
    },
    labelWidth: 90,
  },
  schema: [
    {
      component: "Input",
      componentProps: {
        "data-testid": "sicau-niu-niu-code-input",
        maxlength: 64,
        placeholder: "请输入牛序号",
      },
      fieldName: "code",
      label: "序号",
      rules: z.string().min(1, { message: "请输入牛序号" }),
    },
    {
      component: "Select",
      componentProps: {
        "data-testid": "sicau-niu-niu-type-select",
        options: [
          { label: "普通", value: "common" },
          { label: "特殊", value: "special" },
        ],
      },
      defaultValue: "common",
      fieldName: "niuType",
      label: "类型",
      rules: z.string().min(1, { message: "请选择牛类型" }),
    },
    {
      component: "Select",
      componentProps: {
        "data-testid": "sicau-niu-niu-subtype-select",
        options: [
          { label: "学院", value: "college" },
          { label: "贡献", value: "contribution" },
          { label: "校友", value: "alumni" },
          { label: "精神", value: "spirit" },
        ],
      },
      dependencies: {
        show: (values) => values.niuType === "special",
        triggerFields: ["niuType"],
      },
      fieldName: "specialSubtype",
      label: "特殊子类",
    },
    {
      component: "Input",
      componentProps: {
        "data-testid": "sicau-niu-niu-name-input",
        maxlength: 128,
        placeholder: "请输入牛名称",
      },
      dependencies: {
        show: (values) => values.niuType === "special",
        triggerFields: ["niuType"],
      },
      fieldName: "name",
      label: "名称",
    },
    {
      component: "Select",
      componentProps: {
        "data-testid": "sicau-niu-niu-college-select",
        options: [],
        placeholder: "请选择所属院系",
        showSearch: true,
        optionFilterProp: "label",
      },
      dependencies: {
        show: (values) =>
          values.niuType === "special" && values.specialSubtype === "college",
        triggerFields: ["niuType", "specialSubtype"],
      },
      fieldName: "collegeId",
      label: "所属院系",
    },
    {
      component: "InputNumber",
      componentProps: {
        "data-testid": "sicau-niu-niu-lat-input",
        class: "w-full",
        placeholder: "GPS 纬度",
        precision: 6,
      },
      fieldName: "lat",
      label: "纬度",
    },
    {
      component: "InputNumber",
      componentProps: {
        "data-testid": "sicau-niu-niu-lng-input",
        class: "w-full",
        placeholder: "GPS 经度",
        precision: 6,
      },
      fieldName: "lng",
      label: "经度",
    },
    {
      component: "Select",
      componentProps: {
        "data-testid": "sicau-niu-niu-stage-select",
        allowClear: true,
        options: [
          { label: "预热", value: "warmup" },
          { label: "主体", value: "main" },
          { label: "高潮", value: "climax" },
          { label: "收尾", value: "closing" },
        ],
        placeholder: "请选择放出阶段",
      },
      fieldName: "releaseStage",
      label: "放出阶段",
    },
    {
      component: "DatePicker",
      componentProps: {
        "data-testid": "sicau-niu-niu-online-input",
        class: "w-full",
        showTime: true,
        valueFormat: "x",
      },
      fieldName: "onlineAt",
      label: "上线时间",
    },
    {
      component: "Input",
      componentProps: {
        "data-testid": "sicau-niu-niu-weekdays-input",
        maxlength: 32,
        placeholder: "可选，逗号分隔，如 1,3,5",
      },
      fieldName: "visibleWeekdays",
      label: "可见周几",
    },
    {
      component: "Input",
      componentProps: {
        "data-testid": "sicau-niu-niu-start-input",
        maxlength: 5,
        placeholder: "可选，如 08:00",
      },
      fieldName: "visibleStart",
      label: "可见开始",
    },
    {
      component: "Input",
      componentProps: {
        "data-testid": "sicau-niu-niu-end-input",
        maxlength: 5,
        placeholder: "可选，如 20:00",
      },
      fieldName: "visibleEnd",
      label: "可见结束",
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

async function loadCollegeOptions() {
  const { items } = await listColleges({ pageNum: 1, pageSize: 100 });
  formApi.updateSchema([
    {
      fieldName: "collegeId",
      componentProps: {
        "data-testid": "sicau-niu-niu-college-select",
        options: items.map((item) => ({
          label: item.name,
          value: item.id,
        })),
        placeholder: "请选择所属院系",
        showSearch: true,
        optionFilterProp: "label",
      },
    },
  ]);
}

async function handleConfirm() {
  try {
    modalApi.lock(true);
    const { valid } = await formApi.validate();
    if (!valid) {
      return;
    }
    const values = await formApi.getValues<NiuFormValues>();
    const isSpecial = values.niuType === "special";
    const isCollege = isSpecial && values.specialSubtype === "college";
    const payload = {
      code: values.code.trim(),
      niuType: values.niuType,
      specialSubtype: isSpecial ? values.specialSubtype : "",
      name: isSpecial ? values.name.trim() : "",
      collegeId: isCollege ? (values.collegeId ?? 0) : 0,
      lat: values.lat ?? 0,
      lng: values.lng ?? 0,
      releaseStage: values.releaseStage ?? "",
      onlineAt:
        values.onlineAt === null || values.onlineAt === undefined
          ? null
          : Number(values.onlineAt),
      visibleWeekdays: values.visibleWeekdays?.trim() ?? "",
      visibleStart: values.visibleStart?.trim() ?? "",
      visibleEnd: values.visibleEnd?.trim() ?? "",
    };
    if (isEdit.value) {
      await updateNiu(recordId.value, payload);
      message.success("更新成功");
    } else {
      await createNiu(payload);
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
  await loadCollegeOptions();
  const data = modalApi.getData<Partial<NiuItem>>();
  recordId.value = data?.id ?? 0;
  if (recordId.value > 0) {
    const detail = await getNiu(recordId.value);
    await formApi.setValues({
      code: detail.code,
      niuType: detail.niuType,
      specialSubtype: detail.specialSubtype,
      name: detail.name,
      collegeId: detail.collegeId > 0 ? detail.collegeId : null,
      lat: detail.lat,
      lng: detail.lng,
      releaseStage: detail.releaseStage,
      onlineAt: detail.onlineAt,
      visibleWeekdays: detail.visibleWeekdays,
      visibleStart: detail.visibleStart,
      visibleEnd: detail.visibleEnd,
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
    <div data-testid="sicau-niu-niu-form">
      <NiuForm />
    </div>
  </Modal>
</template>
