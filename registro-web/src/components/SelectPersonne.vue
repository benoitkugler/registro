<template>
  <v-autocomplete
    :label="props.label"
    variant="outlined"
    density="compact"
    :readonly="props.readonly"
    hide-details
    :items="items"
    :model-value="zeroableToNullable(modelValue)"
    @update:model-value="(v) => (modelValue = nullableToZeroable(v))"
    v-model:search="search"
    @update:search="updateSearch"
    clearable
    :no-data-text="
      search.length
        ? 'Aucun profil ne correspond à votre recherche.'
        : 'Veuillez entrer une recherche.'
    "
  ></v-autocomplete>
</template>

<script setup lang="ts">
import type { IdPersonne } from "@/clients/backoffice/logic/api";
import { nullableToZeroable, zeroableToNullable } from "@/utils";
import { ref } from "vue";
import type { SelectPersonneAPI } from "./types";
const props = defineProps<{
  label: string;
  initialPersonne: string;
  api: SelectPersonneAPI;
  readonly?: boolean;
}>();

const modelValue = defineModel<IdPersonne>({ required: true });

const items = ref<{ title: string; value: IdPersonne }[]>(
  modelValue.value == 0
    ? []
    : [{ title: props.initialPersonne, value: modelValue.value }]
);
const search = ref("");
async function doSearch() {
  if (!search.value) return;
  const res = await props.api.SelectPersonne({ search: search.value });
  if (res === undefined) return;
  items.value = (res || []).map((p) => ({ title: p.Label, value: p.Id }));
}

// debounce feature for search
let timerId: ReturnType<typeof setTimeout>;
const debounceDelay = 300;
function updateSearch(s: string | null) {
  search.value = s || "";

  // cancel pending call
  clearTimeout(timerId);

  // delay new call
  timerId = setTimeout(doSearch, debounceDelay);
}
</script>

<style scoped></style>
