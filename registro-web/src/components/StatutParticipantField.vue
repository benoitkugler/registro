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
  StatutParticipant,
  StatutParticipantLabels,
} from "@/clients/backoffice/logic/api";
import { Formatters, selectItems } from "@/utils";
import { computed } from "vue";
const props = defineProps<{
  hideDetails?: boolean;
  readonly?: boolean;
  rules?: any[];
  restrictItems?: StatutParticipant[];
}>();

const modelValue = defineModel<StatutParticipant>({ required: true });

const items = computed(() =>
  selectItems(StatutParticipantLabels)
    .filter(
      (s) =>
        props.restrictItems === undefined ||
        props.restrictItems.includes(s.value)
    )
    .map((statut) => ({
      value: statut.value,
      title: statut.title,
      format: Formatters.statutParticipant(statut.value),
    }))
);
</script>

<style scoped></style>
