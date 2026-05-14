<script setup lang="ts">
import { computed, h, ref, watch } from "vue";

import {
  Button as AButton,
  Segmented as ASegmented,
  Textarea as ATextarea,
} from "ant-design-vue";

import { TiptapEditor } from "#/components/tiptap";
import { $t } from "#/locales";

const props = withDefaults(
  defineProps<{
    disabled?: boolean;
    height?: number | string;
    modelValue?: string;
    scene?: string;
  }>(),
  {
    disabled: false,
    height: 300,
    modelValue: "",
    scene: "other",
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const mode = ref<"source" | "visual">("visual");
const sourceValue = ref(props.modelValue);

const heightStyle = computed(() =>
  typeof props.height === "number" ? `${props.height}px` : props.height,
);
const modeOptions = computed(() => [
  {
    label: h(
      "span",
      { "data-testid": "tiptap-mode-visual" },
      $t("plugin.cms.editor.modes.visual"),
    ),
    value: "visual",
  },
  {
    label: h(
      "span",
      { "data-testid": "tiptap-mode-source" },
      $t("plugin.cms.editor.modes.source"),
    ),
    value: "source",
  },
]);

watch(
  () => props.modelValue,
  (value) => {
    sourceValue.value = value || "";
  },
);

watch(mode, (nextMode) => {
  if (nextMode === "source") {
    sourceValue.value = props.modelValue || "";
  }
});

function updateModelValue(value: string) {
  sourceValue.value = value;
  emit("update:modelValue", value);
}

function applySourceValue() {
  updateModelValue(sourceValue.value || "");
  mode.value = "visual";
}
</script>

<template>
  <div class="cms-rich-editor" :class="{ 'is-disabled': disabled }">
    <div class="cms-rich-editor-modebar">
      <ASegmented
        v-model:value="mode"
        data-testid="tiptap-editor-mode"
        :disabled="disabled"
        :options="modeOptions"
        size="small"
      />
      <AButton
        v-if="mode === 'source'"
        data-testid="tiptap-source-apply"
        size="small"
        type="primary"
        @click="applySourceValue"
      >
        {{ $t("plugin.cms.editor.actions.applySource") }}
      </AButton>
    </div>

    <TiptapEditor
      v-if="mode === 'visual'"
      :model-value="modelValue"
      :disabled="disabled"
      :height="height"
      :placeholder="$t('plugin.cms.editor.placeholder')"
      :scene="scene"
      @update:model-value="updateModelValue"
    />

    <div v-else class="cms-rich-editor-source" :style="{ minHeight: heightStyle }">
      <ATextarea
        v-model:value="sourceValue"
        class="cms-rich-editor-source-input"
        data-testid="tiptap-source-input"
        :disabled="disabled"
        :placeholder="$t('plugin.cms.editor.source.placeholder')"
        @blur="updateModelValue(sourceValue || '')"
      />
      <div
        class="cms-rich-editor-source-preview"
        data-testid="tiptap-source-preview"
        v-html="sourceValue"
      ></div>
    </div>
  </div>
</template>

<style scoped>
.cms-rich-editor {
  overflow: hidden;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
}

.cms-rich-editor:focus-within {
  border-color: #1677ff;
  box-shadow: 0 0 0 2px rgb(5 145 255 / 10%);
}

.cms-rich-editor.is-disabled {
  cursor: not-allowed;
  background: #f5f5f5;
}

.cms-rich-editor :deep(.tiptap-editor) {
  border: 0;
  border-radius: 0;
}

.cms-rich-editor-modebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 6px 8px;
  border-bottom: 1px solid #d9d9d9;
  background: #fff;
}

.cms-rich-editor-source {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  overflow: hidden;
}

.cms-rich-editor-source-input {
  height: 100%;
  min-height: inherit;
  font-family:
    ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono",
    monospace;
  font-size: 12px;
  line-height: 1.6;
  border: 0;
  border-radius: 0;
  resize: none;
}

.cms-rich-editor-source-preview {
  min-height: inherit;
  padding: 12px 16px;
  overflow: auto;
  background: #fff;
  border-left: 1px solid #d9d9d9;
}

.cms-rich-editor-source-preview :deep(img) {
  max-width: 100%;
  height: auto;
}
</style>
