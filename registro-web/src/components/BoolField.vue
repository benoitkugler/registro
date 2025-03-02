<template>
  <v-checkbox
    :label="props.label"
    :model-value="modelValue"
    @update:model-value="(v) => (modelValue = v || false)"
    :hide-details="props.hideDetails"
    :readonly="props.readonly"
    :density="props.density"
    :persistent-hint="props.persistentHint"
    :hint="props.hint"
  ></v-checkbox>
</template>

<script setup lang="ts">
const props = defineProps<{
  label: string;
  hideDetails?: boolean;
  readonly?: boolean;
  density?: "default" | "comfortable" | "compact";
  persistentHint?: boolean;
  hint?: string;
}>();

const modelValue = defineModel<boolean>({ required: true });
</script>

<style scoped></style>
