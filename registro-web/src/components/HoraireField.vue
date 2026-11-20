<template>
  <v-select
    :label="props.label"
    variant="outlined"
    density="compact"
    hide-details
    :items="items"
    v-model="modelValue"
  ></v-select>
</template>

<script setup lang="ts">
import { type Horaire, type Int } from "@/clients/backoffice/logic/api";
const props = defineProps<{
  label: string;
}>();

const modelValue = defineModel<Horaire>({ required: true });

const items = [
  {
    value: { Valid: false, Heure: 0 as Int, Minute: 0 as Int },
    title: "Ne pas préciser",
  },
].concat(
  [8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20].flatMap((heure) =>
    [0, 15, 30, 45].map((min) => ({
      value: {
        Valid: true,
        Heure: heure as Int,
        Minute: min as Int,
      } satisfies Horaire,
      title: `${heure}:${min.toFixed().padStart(2, "0")}`,
    })),
  ),
);
</script>

<style scoped></style>
