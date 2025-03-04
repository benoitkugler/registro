<template>
  <v-select
    label="Mode de paiement"
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
  ModePaiement,
  ModePaiementLabels,
} from "@/clients/backoffice/logic/api";
import { selectItems } from "@/utils";
const props = defineProps<{
  hideDetails?: boolean;
  readonly?: boolean;
  rules?: any[];
}>();

const modelValue = defineModel<ModePaiement>({ required: true });

const items = selectItems(ModePaiementLabels);
</script>

<style scoped></style>
