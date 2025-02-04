<template>
  <v-text-field
    :label="props.label"
    density="compact"
    variant="outlined"
    type="number"
    :min="props.min || 1"
    :max="props.max"
    v-model.number="model"
    hide-details
  ></v-text-field>
</template>

<script setup lang="ts">
import type { Int } from "@/logic/app/api";
const props = defineProps<{
  label: string;
  min?: Int;
  max?: Int;
}>();

const model = defineModel<Int>();
</script>
