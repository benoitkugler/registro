<template>
  <v-card
    title="Bilan"
    subtitle="Sélectionner les séjours à exporter au format Excel."
  >
    <template v-slot:append>
      <v-btn variant="flat" @click="addCurrentYear">
        <template v-slot:prepend><v-icon>mdi-plus</v-icon></template>
        Année {{ props.curentYear }}</v-btn
      >
    </template>
    <v-card-text>
      <v-select
        prepend-icon="mdi-plus"
        placeholder="Rechercher un séjour..."
        :items="selectItems"
        variant="outlined"
        density="comfortable"
        v-model:model-value="currentSearch"
        @update:model-value="
          (v) => {
            toExport.add(v);
            currentSearch = null;
          }
        "
      ></v-select>
      <v-list>
        <div v-if="!selected.length" class="font-italic text-center">
          Aucun séjour n'est encore sélectionné.
        </div>
        <v-list-item
          v-for="(camp, index) in selected"
          :key="index"
          :title="camp.Nom"
          :subtitle="Camps.year(camp)"
        >
          <template v-slot:append>
            <v-tooltip text="Ajouter les autres années">
              <template v-slot:activator="{ isActive, props }">
                <v-btn
                  v-on="{ isActive }"
                  v-bind="props"
                  class="mr-1"
                  icon="mdi-history"
                  size="x-small"
                  @click="addOtherYears(camp)"
                ></v-btn>
              </template>
            </v-tooltip>
            <v-btn icon size="x-small" @click="toExport.delete(camp.Compta.Id)"
              ><v-icon color="red">mdi-close</v-icon></v-btn
            >
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
        :disabled="!toExport.size"
        @click="emit('download', { Ids: Array.from(toExport) })"
      >
        Télécharger
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import type { BilanCampsIn, CampHeader, IdComptaCamp } from "../../logic/api";
import { Camps, normalize } from "@/utils";

const props = defineProps<{
  camps: CampHeader[];
  curentYear: number;
}>();

const emit = defineEmits<{
  (e: "download", selected: BilanCampsIn): void;
}>();

const toExport = ref(new Set<IdComptaCamp>());
const campMap = computed(
  () => new Map(props.camps.map((c) => [c.Compta.Id, c])),
);
const selected = computed(() =>
  Array.from(toExport.value).map((id) => campMap.value.get(id)!),
);

const currentSearch = ref<IdComptaCamp | null>(null);
const selectItems = computed(() =>
  // hide already select camps
  props.camps
    .filter((c) => !toExport.value.has(c.Compta.Id))
    .map((c) => ({
      value: c.Compta.Id,
      title: Camps.label(c),
    })),
);

function addCurrentYear() {
  props.camps
    .filter((c) => Camps.year(c) == props.curentYear)
    .forEach((c) => toExport.value.add(c.Compta.Id));
}

function addOtherYears(camp: CampHeader) {
  searchOtherYears(camp, props.camps).forEach((c) =>
    toExport.value.add(c.Compta.Id),
  );
}

function searchOtherYears(camp: CampHeader, among: CampHeader[]) {
  const query = normalize(camp.Nom);
  const year = Camps.year(camp);
  return among.filter(
    (c) => Camps.year(c) != year && query == normalize(c.Nom),
  );
}
</script>
