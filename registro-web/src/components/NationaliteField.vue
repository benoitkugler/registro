<template>
  <v-select
    label="Nationalité"
    variant="outlined"
    density="compact"
    :items="items"
    :readonly="props.readonly"
    :hide-details="props.hideDetails"
    v-model="modelValue"
    :rules="props.rules"
  ></v-select>
</template>

<script setup lang="ts">
import {
  Nationnalite,
  NationnaliteLabels,
} from "@/clients/inscription/logic/api";
import { selectItems } from "@/utils";
const props = defineProps<{
  hideDetails?: boolean;
  readonly?: boolean;
  rules?: any[];
}>();

const modelValue = defineModel<Nationnalite>({ required: true });

const items = selectItems(NationnaliteLabels);
</script>

<style scoped></style>
