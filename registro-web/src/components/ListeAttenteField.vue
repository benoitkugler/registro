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
      <v-list-item v-bind="menuProps" :title="item.title">
        <template #prepend>
          <v-icon :color="item.raw.format.color">{{
            item.raw.format.icon
          }}</v-icon>
        </template>
      </v-list-item>
    </template>

    <template #selection="{ item }">
      <v-list-item :title="item.title" density="compact">
        <template #prepend>
          <v-icon :color="item.raw.format.color" size="small">{{
            item.raw.format.icon
          }}</v-icon>
        </template>
      </v-list-item>
    </template>
  </v-select>
</template>

<script setup lang="ts">
import {
  ListeAttente,
  ListeAttenteLabels,
} from "@/clients/backoffice/logic/api";
import { Formatters, selectItems } from "@/utils";
const props = defineProps<{
  hideDetails?: boolean;
  readonly?: boolean;
  rules?: any[];
}>();

const modelValue = defineModel<ListeAttente>({ required: true });

const items = selectItems(ListeAttenteLabels).map((statut) => ({
  value: statut.value,
  title: statut.title,
  format: Formatters.listeAttente(statut.value),
}));
</script>

<style scoped></style>
