<template>
  <v-text-field
    :label="props.label"
    density="compact"
    variant="outlined"
    type="number"
    :min="props.min === undefined ? 1 : props.min"
    :max="props.max"
    v-model.number="model"
    :hide-details="!hint"
    :hint="props.hint"
    :persistent-hint="props.persistentHint"
    :disabled="props.disabled"
    :suffix="props.suffix"
  ></v-text-field>
</template>

<script setup lang="ts">
import type { Int } from "@/clients/backoffice/logic/api";
const props = defineProps<{
  label: string;
  hint?: string;
  min?: Int;
  max?: Int;
  disabled?: boolean;
  suffix?: string;
  persistentHint?: boolean;
}>();

const model = defineModel<Int>({ required: true });
</script>
