<template>
  <v-row no-gutters class="ma-2">
    <!-- liste de recherche -->
    <v-col cols="6">
      <v-card
        title="Dossiers"
        :subtitle="
          data == null ? '-' : `${data.Dossiers?.length || 0} / ${data.Total}`
        "
      >
        <template v-slot:append>
          <DebounceField
            width="350"
            density="compact"
            variant="outlined"
            prepend-inner-icon="mdi-magnify"
            label="Rechercher"
            hide-details
            clearable
            v-model="query.Pattern"
            @update:model-value="searchDossiers"
          >
            <template v-slot:append>
              <v-tooltip text="Critères de recherche">
                <template v-slot:activator="{ props: tooltipProps }">
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
        </template>
        <v-expand-transition>
          <v-row v-show="showDetailsQuery" class="mx-0">
            <v-col cols="5">
              <v-select
                density="compact"
                variant="outlined"
                label="Camp"
                :items="campItems"
                clearable
                :model-value="optToNullable(query.IdCamp)"
                @update:model-value="
                  (id) => {
                    query.IdCamp = nullableToOpt(id);
                    searchDossiers();
                  }
                "
              >
                <template v-slot:prepend>
                  <v-tooltip
                    :text="
                      showTerminatedCamps
                        ? 'Masquer les camps terminés'
                        : 'Afficher les camps terminés'
                    "
                  >
                    <template v-slot:activator="{ props: tooltipProps }">
                      <v-btn
                        size="small"
                        :variant="showTerminatedCamps ? 'tonal' : 'flat'"
                        v-bind="tooltipProps"
                        icon="mdi-history"
                        @click="showTerminatedCamps = !showTerminatedCamps"
                      ></v-btn>
                    </template>
                  </v-tooltip>
                </template>
              </v-select>
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
        <v-card-text>
          <v-list v-if="data != null">
            <v-list-item v-if="!data.Dossiers?.length" class="text-center">
              <i>Aucun dossier ne correspond à votre recherche.</i>
            </v-list-item>
            <v-list-item
              v-for="(dossier, i) in data.Dossiers"
              :key="i"
              :title="dossier.Responsable"
              :subtitle="dossier.Participants"
              @click="loadDossier(dossier.Id)"
            >
              <template v-slot:append>
                <v-badge
                  color="primary"
                  :content="dossier.NewMessages"
                ></v-badge>
              </template>
            </v-list-item>
          </v-list>
        </v-card-text>
      </v-card>
    </v-col>
    <!-- pannel de détails -->
    <v-col>
      <DossierDetailsPannel
        v-if="dossierDetails != null"
        :dossier="dossierDetails"
      ></DossierDetailsPannel>
      <div v-else class="text-center font-italic my-6">
        Sélectionner un dossier...
      </div>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import {
  QueryAttente,
  QueryAttenteLabels,
  QueryReglement,
  QueryReglementLabels,
  type CampItem,
  type SearchDossierIn,
  type Int,
  type SearchDossierOut,
  type IdDossier,
  type Personne,
  type DossierDetails,
} from "../../logic/api";
import { copy, nullableToOpt, optToNullable, selectItems } from "@/utils";
import { controller } from "../../logic/logic";
import DebounceField from "@/components/DebounceField.vue";
import DossierDetailsPannel from "./DossierDetailsPannel.vue";
const props = defineProps<{}>();

defineExpose({ showDossier });

const showDetailsQuery = ref(false);

const emptyQuery: SearchDossierIn = {
  Pattern: "",
  IdCamp: { Valid: false, Id: 0 as Int },
  Attente: QueryAttente.EmptyQA,
  Reglement: QueryReglement.EmptyQR,
};

const query = reactive<SearchDossierIn>(copy(emptyQuery));

onMounted(() => {
  loadCamps();
  searchDossiers();
});

const showTerminatedCamps = ref(false);
const campsData = ref<CampItem[]>([]);
const campItems = computed(() =>
  campsData.value
    .filter((c) => (showTerminatedCamps.value ? true : !c.IsOld))
    .map((c) => ({ title: c.Label, value: c.Id }))
);
async function loadCamps() {
  const res = await controller.GetCamps();
  if (res === undefined) return;
  campsData.value = res || [];
}

const data = ref<SearchDossierOut | null>(null);
async function searchDossiers() {
  const res = await controller.DossiersSearch(query);
  if (res === undefined) return;
  data.value = res;
}

// showDossier may be called by the parent when switching to this pannel
function showDossier(id: IdDossier, responsable: Personne) {
  // reset the query so that the given Dossier is found
  query.IdCamp = emptyQuery.IdCamp;
  query.Attente = emptyQuery.Attente;
  query.Reglement = emptyQuery.Reglement;

  query.Pattern = `${responsable.Prenom} ${responsable.Nom}`;
  searchDossiers();
  loadDossier(id);
}

const dossierDetails = ref<DossierDetails | null>(null);
// fetch the complete information and displays it in the right pannel
async function loadDossier(id: IdDossier) {
  const res = await controller.DossiersLoad({ id: id });
  if (res === undefined) return;
  dossierDetails.value = res;
}
</script>
