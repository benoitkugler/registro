<template>
  <v-select
    label="Trajet en navette"
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
import { NavetteLabels, type Navette } from "@/clients/backoffice/logic/api";
import { selectItems } from "@/utils";
const props = defineProps<{
  hideDetails?: boolean;
  readonly?: boolean;
  rules?: any[];
}>();

const modelValue = defineModel<Navette>({ required: true });

const items = selectItems(NavetteLabels);
</script>

<style scoped></style>
