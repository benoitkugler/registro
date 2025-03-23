<template>
  <v-select
    label="Diplôme"
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
import { Diplome, DiplomeLabels } from "@/clients/equipier/logic/api";
import { selectItems } from "@/utils";
const props = defineProps<{
  hideDetails?: boolean;
  readonly?: boolean;
  rules?: any[];
}>();

const modelValue = defineModel<Diplome>({ required: true });

const items = selectItems(DiplomeLabels);
</script>
