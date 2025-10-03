<template>
  <v-checkbox
    label="Nationalité Suisse ?"
    variant="outlined"
    density="compact"
    :readonly="props.readonly"
    hint="Cochez si vous avez la nationalité suisse."
    persistent-hint
    :model-value="modelValue.IsSuisse"
    @update:model-value="(b) => (modelValue.IsSuisse = b || false)"
  ></v-checkbox>
</template>

<script setup lang="ts">
import type { Nationnalite } from "@/clients/inscription/logic/api";
const props = defineProps<{
  readonly?: boolean;
}>();

const modelValue = defineModel<Nationnalite>({ required: true });
</script>

<style scoped></style>
