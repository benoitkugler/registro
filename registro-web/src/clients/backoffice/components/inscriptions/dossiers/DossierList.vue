<template>
  <div>
    <v-row>
      <v-col align-self="center">
        <DebounceField
          density="compact"
          variant="outlined"
          prepend-inner-icon="mdi-magnify"
          label="Rechercher un dossier"
          hide-details
          clearable
          v-model="query.Pattern"
          @update:model-value="searchDossiers"
        >
          <template #append>
            <v-tooltip text="Trier selon les nouveaux messages">
              <template #activator="{ props: tooltipProps }">
                <v-btn
                  v-bind="tooltipProps"
                  icon
                  size="small"
                  class="mr-1"
                  @click="searchMessages"
                >
                  <v-icon color="light-blue-darken-3">mdi-email</v-icon>
                </v-btn>
              </template>
            </v-tooltip>

            <v-tooltip text="Autres critères de recherche">
              <template #activator="{ props: tooltipProps }">
                <v-btn
                  size="small"
                  variant="flat"
                  v-bind="tooltipProps"
                  :icon="
                    showDetailsQuery ? 'mdi-chevron-up' : 'mdi-chevron-down'
                  "
                  @click="showDetailsQuery = !showDetailsQuery"
                ></v-btn>
              </template>
            </v-tooltip>
          </template>
        </DebounceField>
      </v-col>
    </v-row>
    <v-expand-transition>
      <v-row v-show="showDetailsQuery" class="mx-0">
        <v-col cols="5">
          <SelectCamp
            label="Camp"
            :camps="props.camps"
            :model-value="optToNullable(query.IdCamp)"
            @update:model-value="
              (id) => {
                query.IdCamp = nullableToOpt(id);
                searchDossiers();
              }
            "
          ></SelectCamp>
        </v-col>
        <v-col>
          <v-select
            density="compact"
            variant="outlined"
            label="Liste d'attente"
            v-model="query.Attente"
            :items="selectItems(QueryAttenteLabels)"
            @update:model-value="searchDossiers"
          ></v-select>
        </v-col>
        <v-col>
          <v-select
            density="compact"
            variant="outlined"
            label="Réglement"
            v-model="query.Reglement"
            :items="selectItems(QueryReglementLabels)"
            @update:model-value="searchDossiers"
          ></v-select>
        </v-col>
      </v-row>
    </v-expand-transition>
    <v-list v-if="dossierHeaders != null" class="mt-2">
      <v-list-item v-if="!dossierHeaders.Dossiers?.length" class="text-center">
        <i>Aucun dossier ne correspond à votre recherche.</i>
      </v-list-item>
      <v-list-item
        rounded
        v-for="(dossier, i) in dossierHeaders.Dossiers"
        :key="i"
        :title="dossier.Responsable"
        :subtitle="dossier.Participants"
        :value="dossier.Id"
        @click="emit('click', dossier)"
        class="my-1"
      >
        <template #append v-if="dossier.NewMessages">
          <v-badge
            color="light-blue-darken-3"
            :content="dossier.NewMessages"
          ></v-badge>
        </template>
      </v-list-item>
    </v-list>
  </div>
</template>

<script setup lang="ts">
import {
  QueryAttenteLabels,
  QueryReglementLabels,
  type CampItem,
  type DossierHeader,
  type SearchDossierIn,
  type SearchDossierOut,
} from "@/clients/backoffice/logic/api";
import { controller } from "@/clients/backoffice/logic/logic";
import { nullableToOpt, optToNullable, selectItems } from "@/utils";
import { ref, watch } from "vue";

const props = defineProps<{
  camps: CampItem[];
}>();

const emit = defineEmits<{
  (e: "click", dossier: DossierHeader): void;
  (e: "update", dossiers: SearchDossierOut): void;
}>();

const query = defineModel<SearchDossierIn>("query", { required: true });

defineExpose({ searchDossiers });

watch(() => query.value, searchDossiers, { immediate: true });

const showDetailsQuery = ref(false);

const dossierHeaders = ref<SearchDossierOut | null>(null);
async function searchDossiers() {
  const res = await controller.DossiersSearch(query.value);
  if (res === undefined) return;
  dossierHeaders.value = res;
  emit("update", res);
}

function searchMessages() {
  query.value.Pattern = "sort:messages";
  searchDossiers();
}
</script>
