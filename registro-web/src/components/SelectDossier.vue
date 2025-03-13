<template>
  <v-autocomplete
    label="Dossier"
    variant="outlined"
    density="compact"
    :readonly="props.readonly"
    hide-details
    :items="items"
    no-filter
    :model-value="zeroableToNullable(modelValue)"
    @update:model-value="(v) => (modelValue = nullableToZeroable(v))"
    v-model:search="search"
    @update:search="updateSearch"
    clearable
    :no-data-text="
      search.length
        ? 'Aucun dossier ne correspond à votre recherche.'
        : 'Veuillez entrer une recherche.'
    "
  >
    <template #item="{ item, props: menuProps }">
      <v-list-item
        v-bind="menuProps"
        :title="item.title"
        :subtitle="item.raw.subtitle"
      ></v-list-item>
    </template>
  </v-autocomplete>
</template>

<script setup lang="ts">
import type { DossierHeader, IdDossier } from "@/clients/backoffice/logic/api";
import {
  controller,
  emptyQuery,
  idQuery,
} from "@/clients/backoffice/logic/logic";
import { nullableToZeroable, zeroableToNullable } from "@/utils";
import { onMounted, ref } from "vue";
const props = defineProps<{
  readonly?: boolean;
}>();

const modelValue = defineModel<IdDossier>({ required: true });

const items = ref<{ title: string; subtitle: string; value: IdDossier }[]>([]);

onMounted(fetchInitial);

function dossierToItem(d: DossierHeader) {
  return {
    title: d.Responsable,
    subtitle: d.Participants,
    value: d.Id,
  };
}

// make sure the items contains the initial value
async function fetchInitial() {
  const id = modelValue.value;
  if (id == 0) return;
  const res = await controller.DossiersSearch(idQuery(id));
  if (res === undefined) return;
  items.value = (res.Dossiers || []).map(dossierToItem);
}

const search = ref("");
async function doSearch() {
  if (!search.value) return;
  const query = emptyQuery();
  query.Pattern = search.value;
  const res = await controller.DossiersSearch(query);
  if (res === undefined) return;
  items.value = (res.Dossiers || []).map(dossierToItem);
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
