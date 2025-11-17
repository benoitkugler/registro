<template>
  <v-combobox
    variant="outlined"
    density="compact"
    v-model="modelValue"
    :label="props.label"
    hint="Ou pays si hors de France"
    :rules="props.rules"
    :items="Departements"
  ></v-combobox>
</template>

<script setup lang="ts">
import type { Departement } from "@/clients/backoffice/logic/api";
import { Departements } from "@/utils";

const props = defineProps<{
  label: string;
  rules?: [(v: string) => true | string];
  readonly?: boolean;
}>();

const modelValue = defineModel<Departement>({ required: true });
</script>

<style scoped></style>
