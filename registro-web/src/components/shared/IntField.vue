<template>
  <v-text-field
    :label="props.label"
    density="compact"
    variant="outlined"
    type="number"
    :min="props.min || 1"
    :max="props.max"
    v-model.number="model"
    :hide-details="!hint"
    :hint="props.hint"
    :disabled="props.disabled"
  ></v-text-field>
</template>

<script setup lang="ts">
import type { Int } from "@/logic/backoffice/api";
const props = defineProps<{
  label: string;
  hint?: string;
  min?: Int;
  max?: Int;
  disabled?: boolean;
}>();

const model = defineModel<Int>();
</script>
