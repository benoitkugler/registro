<template>
  <v-select
    label="État civil"
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
import { EtatCivil, EtatCivilLabels } from "@/clients/equipier/logic/api";
import { selectItems } from "@/utils";
const props = defineProps<{
  hideDetails?: boolean;
  readonly?: boolean;
  rules?: any[];
}>();

const modelValue = defineModel<EtatCivil>({ required: true });

const items = selectItems(EtatCivilLabels);
</script>
