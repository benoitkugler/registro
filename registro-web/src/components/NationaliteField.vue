<template>
  <v-checkbox
    label="Nationalité Suisse ?"
    variant="outlined"
    density="compact"
    :readonly="props.readonly"
    :hint="props.hint"
    :persistent-hint="!!props.hint"
    :model-value="modelValue.IsSuisse"
    @update:model-value="(b) => (modelValue.IsSuisse = b || false)"
  ></v-checkbox>
</template>

<script setup lang="ts">
import type { Nationnalite } from "@/clients/inscription/logic/api";
const props = defineProps<{
  hint?: string;
  readonly?: boolean;
}>();

const modelValue = defineModel<Nationnalite>({ required: true });
</script>

<style scoped></style>
