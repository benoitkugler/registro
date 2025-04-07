<template>
  <v-select
    chips
    closable-chips
    label="Rôles"
    variant="outlined"
    density="compact"
    hide-details
    :items="items"
    :readonly="props.readonly"
    multiple
    v-model="modelValue"
  ></v-select>
</template>

<script setup lang="ts">
import { selectItems } from "@/utils";
import { RoleLabels, type Roles } from "../../logic/api";

const props = defineProps<{
  readonly?: boolean;
}>();

const modelValue = defineModel<Roles>({ required: true });

const items = selectItems(RoleLabels);
</script>

<style scoped></style>
