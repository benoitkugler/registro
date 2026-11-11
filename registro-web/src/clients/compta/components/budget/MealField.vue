<template>
  <v-select
    variant="outlined"
    density="compact"
    :items="selectItems(MealLabels)"
    :label="props.label"
    :hint="props.hint"
    v-model="model"
    hide-details
  ></v-select>
</template>

<script lang="ts" setup>
import { Meal, MealLabels } from "../../logic/api";
import { selectItems } from "@/utils";

const props = defineProps<{
  label: string;
  hint?: string;
}>();

const model = defineModel<Meal>({ required: true });
</script>
