<template>
  <v-select
    label="Statut"
    variant="outlined"
    density="compact"
    :items="items"
    :readonly="props.readonly"
    :hide-details="props.hideDetails"
    v-model="modelValue"
    :rules="props.rules"
  >
    <template #item="{ item, props: menuProps }">
      <v-list-item
        v-bind="menuProps"
        :title="item.title"
        :prepend-icon="item.raw.icon"
      ></v-list-item>
    </template>
  </v-select>
</template>

<script setup lang="ts">
import {
  ListeAttente,
  ListeAttenteLabels,
} from "@/clients/backoffice/logic/api";
import { selectItems } from "@/utils";
const props = defineProps<{
  hideDetails?: boolean;
  readonly?: boolean;
  rules?: any[];
}>();

const modelValue = defineModel<ListeAttente>({ required: true });

const items = selectItems(ListeAttenteLabels).map((statut) => ({
  value: statut.value,
  title: statut.title,
  icon: statut.value == ListeAttente.Inscrit ? "mdi-check" : "mdi-clock",
}));
</script>

<style scoped></style>
