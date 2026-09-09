<template>
  <v-select
    label="Diplôme spécifique animation"
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

const asso = import.meta.env.VITE_ASSO;

// Keep in sync with the server Diplome enum

const plage =
  asso == "acve"
    ? ([Diplome.DAcveBafa, Diplome.DAcveBeatep] as const)
    : ([Diplome.DRepereJsMoniteur, Diplome.DRepereForje] as const);

const items = selectItems(DiplomeLabels).filter(
  (item) =>
    item.value == Diplome.DAucun ||
    item.value == Diplome.DAutre ||
    (plage[0] <= item.value && item.value <= plage[1]),
);
</script>
