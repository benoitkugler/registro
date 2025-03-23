<template>
  <v-select
    label="Approfondissement"
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
  Approfondissement,
  ApprofondissementLabels,
} from "@/clients/equipier/logic/api";
import { selectItems } from "@/utils";
const props = defineProps<{
  hideDetails?: boolean;
  readonly?: boolean;
  rules?: any[];
}>();

const modelValue = defineModel<Approfondissement>({ required: true });

const items = selectItems(ApprofondissementLabels);
</script>
