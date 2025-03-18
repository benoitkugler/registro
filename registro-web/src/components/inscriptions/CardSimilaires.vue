<template>
  <v-card
    title="Identifier le profil"
    width="600px"
    :subtitle="showManualSearch ? '' : 'Liste des profils similaires'"
  >
    <v-card-text class="pb-0">
      <!-- manual mode -->
      <div v-if="showManualSearch">
        <DebounceField
          v-model="manualPattern"
          @update:model-value="manualSearch"
          label="Rechercher un profil"
          density="comfortable"
          clearable
          autofocus
          hide-details
        ></DebounceField>
        <v-list density="compact">
          <v-row
            class="ma-2"
            no-gutters
            v-if="manualPattern.length >= 3 && !manualSearchCandidates.length"
          >
            <v-col class="text-center">
              <i>Aucun profil ne correspond à votre recherche.</i>
            </v-col>
          </v-row>

          <v-list-item
            v-for="(pers, i) in manualSearchCandidates"
            :key="i"
            :title="pers.Label"
            :subtitle="Formatters.dateNaissance(pers.DateNaissance)"
            :append-icon="Formatters.sexeIcon(pers.Sexe)"
            @click="rattacheTo(pers.Id)"
          >
          </v-list-item>
        </v-list>
      </div>
      <!-- automatic mode (suggestion from server) -->
      <v-list density="compact" v-else>
        <v-row class="mx-2" no-gutters v-if="!suggestedCandidats.length">
          <v-col class="text-center">
            <i>Aucun profil existant sur la base n'est similaire.</i>
          </v-col>
        </v-row>

        <v-list-item
          v-for="(pers, i) in suggestedCandidats"
          :key="i"
          :title="pers.Personne.Label"
          :subtitle="Formatters.dateNaissance(pers.Personne.DateNaissance)"
          :prepend-icon="Formatters.sexeIcon(pers.Personne.Sexe)"
          @click="rattacheTo(pers.Personne.Id)"
        >
          <template #append>
            <v-chip :color="pertinenceColor(pers.ScorePercent)">
              {{ pers.ScorePercent }} %
            </v-chip>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>
    <v-card-actions>
      <v-btn
        @click="
          showManualSearch = !showManualSearch;
          manualPattern = '';
        "
      >
        {{
          showManualSearch ? "Retour aux suggestions" : "Chercher manuellement"
        }}
      </v-btn>
      <v-spacer></v-spacer>
      <v-btn variant="elevated" elevation="1" @click="toNewProfile()">
        <template #prepend>
          <v-icon>mdi-plus</v-icon>
        </template>
        Créer un nouveau profil</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import type {
  IdentTarget,
  IdPersonne,
  Personne,
  PersonneHeader,
  ScoredPersonne,
} from "../../clients/backoffice/logic/api";
import { Formatters } from "@/utils";
import type { SimilairesAPI } from "./types";

const props = defineProps<{
  personne: Personne;
  api: SimilairesAPI;
}>();

const emit = defineEmits<{
  (e: "identifie", params: IdentTarget): void;
}>();

onMounted(fetchSimilaires);

const suggestedCandidats = ref<ScoredPersonne[]>([]);
async function fetchSimilaires() {
  const res = await props.api.searchSimilaires({
    idPersonne: props.personne.Id,
  });
  if (res === undefined) return;
  suggestedCandidats.value = res || [];
}

const showManualSearch = ref(false);

const manualPattern = ref("");
const manualSearchCandidates = ref<PersonneHeader[]>([]);
async function manualSearch() {
  if ((manualPattern.value || "").length < 3) return;
  const res = await props.api.selectPersonne({ search: manualPattern.value });
  if (res === undefined) return;
  manualSearchCandidates.value = res || [];
}

function toNewProfile() {
  emit("identifie", {
    IdTemporaire: props.personne.Id,
    Rattache: false,
    RattacheTo: 0 as IdPersonne,
  });
}

function rattacheTo(id: IdPersonne) {
  emit("identifie", {
    IdTemporaire: props.personne.Id,
    Rattache: true,
    RattacheTo: id,
  });
}

function pertinenceColor(percent: number) {
  const green = Math.floor(percent * 2);
  const color = `rgb(${255 - green},${green}, 15)`;
  return color;
}
</script>
