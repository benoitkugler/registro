<template>
  <v-autocomplete
    multiple
    chips
    closable-chips
    :items="items"
    v-model="model"
  ></v-autocomplete>
</template>

<script setup lang="ts">
import { type CampItem, type IdCamp } from "@/clients/backoffice/logic/api";
import { computed } from "vue";
const props = defineProps<{
  allCamps: CampItem[];
}>();
const emit = defineEmits<{}>();
const model = defineModel<IdCamp[]>({ required: true });

const items = computed(() =>
  props.allCamps.map((c) => ({ title: c.Label, value: c.Id }))
);
</script>
