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
  >
    <template #no-data>
      <v-list-item
        :title="
          search.trim().length
            ? 'Aucun profil ne correspond à votre recherche.'
            : 'Veuillez entrer une recherche.'
        "
      >
        <template
          #append
          v-if="props.enableCreate && !props.readonly && search.trim().length"
        >
          <v-btn size="small" prepend-icon="mdi-plus" @click="emit('create')"
            >Créer...</v-btn
          >
        </template>
      </v-list-item>
    </template>
    <template #item="{ item, props: innerProps }">
      <v-list-item
        v-bind="innerProps"
        :title="item.title"
        :subtitle="`ID : ${item.value}`"
        :value="item.value"
      ></v-list-item>
    </template>
  </v-autocomplete>
</template>

<script setup lang="ts">
import type { IdPersonne } from "@/clients/backoffice/logic/api";
import { nullableToZeroable, Personnes, zeroableToNullable } from "@/utils";
import { ref } from "vue";
import type { SelectPersonneAPI } from "./types";
const props = defineProps<{
  label: string;
  initialPersonne: string;
  api: SelectPersonneAPI;
  readonly?: boolean;
  enableCreate?: boolean;
}>();

const modelValue = defineModel<IdPersonne>({ required: true });

const emit = defineEmits<{
  (e: "create"): void;
}>();

defineExpose({ setWithTitle });

const items = ref<{ title: string; value: IdPersonne }[]>(
  modelValue.value
    ? [{ title: props.initialPersonne, value: modelValue.value }]
    : [],
);
const search = ref("");
async function doSearch() {
  if (!search.value) return;
  const res = await props.api.SelectPersonne({ search: search.value });
  if (res === undefined) return;
  items.value = (res || []).map((p) => ({
    title: Personnes.label(p),
    value: p.Id,
  }));
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

function setWithTitle(id: IdPersonne, title: string) {
  items.value = [{ title, value: id }];
  modelValue.value = id;
}
</script>

<style scoped></style>
