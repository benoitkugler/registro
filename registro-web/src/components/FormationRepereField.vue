<template>
  <v-select
    label="Formation équivalente"
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
  FormationRepereLabels,
  type FormationRepere,
} from "@/clients/equipier/logic/api";
import { selectItems } from "@/utils";
const props = defineProps<{
  hideDetails?: boolean;
  readonly?: boolean;
  rules?: any[];
}>();

const modelValue = defineModel<FormationRepere>({ required: true });

const items = selectItems(FormationRepereLabels);
</script>
