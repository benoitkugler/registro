<template>
  <v-select
    label="Sexe"
    variant="outlined"
    density="compact"
    :items="items"
    :readonly="props.readonly"
    :hide-details="props.hideDetails"
    v-model="modelValue"
    :error-messages="props.error ? [props.error] : null"
  ></v-select>
</template>

<script setup lang="ts">
import { Sexe, SexeLabels } from "@/clients/inscription/logic/api";
import { selectItems } from "@/utils";
const props = defineProps<{
  hideDetails?: boolean;
  readonly?: boolean;
  error?: string;
}>();

const modelValue = defineModel<Sexe>({ required: true });

const items = selectItems(SexeLabels).filter((v) => v.value != Sexe.Empty);
</script>

<style scoped></style>
