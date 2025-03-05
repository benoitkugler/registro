<template>
  <v-row no-gutters class="ma-2">
    <!-- create dossier dialog -->
    <v-dialog v-model="showCreateDossier" max-width="400px">
      <CreateDossierCard @create="createDossier"></CreateDossierCard>
    </v-dialog>
    <!-- liste de recherche -->
    <v-col cols="6">
      <v-card
        title="Dossiers"
        :subtitle="
          dossierHeaders == null
            ? '-'
            : `${dossierHeaders.Dossiers?.length || 0} / ${
                dossierHeaders.Total
              }`
        "
      >
        <template #append>
          <v-row>
            <v-col align-self="center">
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
                <template #append>
                  <v-tooltip text="Critères de recherche">
                    <template #activator="{ props: tooltipProps }">
                      <v-btn
                        size="small"
                        variant="flat"
                        v-bind="tooltipProps"
                        :icon="
                          showDetailsQuery
                            ? 'mdi-chevron-up'
                            : 'mdi-chevron-down'
                        "
                        @click="showDetailsQuery = !showDetailsQuery"
                      ></v-btn>
                    </template>
                  </v-tooltip>
                </template>
              </DebounceField>
            </v-col>
            <v-col align-self="center">
              <v-btn icon size="small" @click="showCreateDossier = true">
                <v-icon color="green">mdi-plus</v-icon>
              </v-btn>
            </v-col>
          </v-row>
        </template>
        <v-expand-transition>
          <v-row v-show="showDetailsQuery" class="mx-0">
            <v-col cols="5">
              <SelectCamp
                label="Camp"
                :camps="allCamps"
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
        <v-card-text>
          <v-list v-if="dossierHeaders != null">
            <v-list-item
              v-if="!dossierHeaders.Dossiers?.length"
              class="text-center"
            >
              <i>Aucun dossier ne correspond à votre recherche.</i>
            </v-list-item>
            <v-list-item
              v-for="(dossier, i) in dossierHeaders.Dossiers"
              :key="i"
              :title="dossier.Responsable"
              :subtitle="dossier.Participants"
              @click="loadDossier(dossier.Id)"
            >
              <template #append v-if="dossier.NewMessages">
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
        :structures="structures"
        :camps="allCamps"
        @update-dossier="updateDossier"
        @delete-dossier="deleteDossier"
        @create-participant="createParticipant"
        @update-participant="updateParticipant"
        @delete-participant="deleteParticipant"
        @expand-participant="expandParticipant"
        @create-aide="createAide"
        @delete-aide="deleteAide"
        @update-aide="updateAide"
        @create-paiement="createPaiement"
        @update-paiement="updatePaiement"
        ref="detailsPannel"
      ></DossierDetailsPannel>
      <div v-else class="text-center font-italic my-6">
        Sélectionner un dossier...
      </div>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, useTemplateRef } from "vue";
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
  type DossierDetails,
  type Dossier,
  type AidesCreateIn,
  type Structureaides,
  type IdAide,
  type Aide,
  type Participant,
  type ParticipantsCreateIn,
  type IdParticipant,
  type Paiement,
  type IdPersonne,
} from "../../logic/api";
import { copy, nullableToOpt, optToNullable, selectItems } from "@/utils";
import { controller } from "../../logic/logic";
import DebounceField from "@/components/DebounceField.vue";
import DossierDetailsPannel from "./dossiers/DossierDetailsPannel.vue";
import CreateDossierCard from "./dossiers/CreateDossierCard.vue";

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
  loadStructureaides();
  searchDossiers();
});

const detailsPannel = useTemplateRef("detailsPannel");

const allCamps = ref<CampItem[]>([]);
async function loadCamps() {
  const res = await controller.GetCamps();
  if (res === undefined) return;
  allCamps.value = res || [];
}

const structures = ref<NonNullable<Structureaides>>({});
async function loadStructureaides() {
  const res = await controller.GetStructureaides();
  if (res === undefined) return;
  structures.value = res || {};
}

const dossierHeaders = ref<SearchDossierOut | null>(null);
async function searchDossiers() {
  const res = await controller.DossiersSearch(query);
  if (res === undefined) return;
  dossierHeaders.value = res;

  // if no dossier is yet selected, auto select the first
  if (dossierDetails.value == null && res.Dossiers?.length) {
    loadDossier(res.Dossiers[0].Id);
  }
}

// showDossier may be called by the parent when switching to this pannel;
//  it loads the dossier details
async function showDossier(id: IdDossier, responsable: string) {
  // reset the query so that the given Dossier is found
  query.IdCamp = emptyQuery.IdCamp;
  query.Attente = emptyQuery.Attente;
  query.Reglement = emptyQuery.Reglement;
  query.Pattern = responsable;
  searchDossiers();
  await loadDossier(id);
}

const dossierDetails = ref<DossierDetails | null>(null);
// fetch the complete information and displays it in the right pannel
async function loadDossier(id: IdDossier) {
  const res = await controller.DossiersLoad({ id: id });
  if (res === undefined) return;
  dossierDetails.value = res;
}

// appelée si besoin de mettre à jour l'état financier
async function ensureDossier() {
  if (dossierDetails.value != null)
    loadDossier(dossierDetails.value.Dossier.Dossier.Id);
}

const showCreateDossier = ref(false);
async function createDossier(idResponsable: IdPersonne) {
  showCreateDossier.value = false;
  const res = await controller.DossiersCreate({ idResponsable });
  if (res === undefined) return;
  controller.showMessage("Dossier créé avec succès.");
  // add to the list and select it
  await showDossier(res.Id, res.Responsable);
  // also start editing
  detailsPannel.value?.showEditDossier();
}

async function updateDossier(dossier: Dossier) {
  const res = await controller.DossiersUpdate(dossier);
  if (res === undefined) return;
  controller.showMessage("Dossier mis à jour avec succès.");

  // properly update current display
  if (dossierDetails.value != null) {
    dossierDetails.value.Dossier.Dossier = dossier;
    dossierDetails.value.Dossier.Responsable = res.Responsable;
  }
  const dossierHeader = dossierHeaders.value?.Dossiers?.find(
    (d) => d.Id == dossier.Id
  );
  if (dossierHeader) dossierHeader.Responsable = res.Responsable;
}

async function deleteDossier() {
  if (dossierDetails.value == null) return;
  const toDelete = dossierDetails.value.Dossier.Dossier.Id;
  const res = await controller.DossiersDelete({
    id: toDelete,
  });
  if (res === undefined) return;

  controller.showMessage("Dossier supprimé avec succès.");
  // properly cleanup
  dossierDetails.value = null;
  searchDossiers();
}

async function createParticipant(args: ParticipantsCreateIn) {
  const res = await controller.ParticipantsCreate(args);
  if (res === undefined) return;

  controller.showMessage("Participant ajouté avec succès.");
  ensureDossier();
}

async function updateParticipant(participant: Participant) {
  const res = await controller.ParticipantsUpdate(participant);
  if (res === undefined) return;

  controller.showMessage("Réglages du participant modifiés avec succès.");
  ensureDossier();
}

async function deleteParticipant(id: IdParticipant) {
  const res = await controller.ParticipantsDelete({ id });
  if (res === undefined) return;

  controller.showMessage("Participant supprimé avec succès.");
  ensureDossier();
}

/** copy the [Remises] and [OptionPrix] (for same camp) fields */
async function expandParticipant(participant: Participant) {
  if (dossierDetails.value == null) return;
  const others =
    dossierDetails.value.Dossier.Participants?.map((p) => p.Participant).filter(
      (p) => p.Id != participant.Id
    ) || [];
  if (!others.length) return;

  // copy the fields
  others.forEach((p) => {
    p.Remises = participant.Remises;
    p.QuotientFamilial = participant.QuotientFamilial;
    if (p.IdCamp == participant.IdCamp) {
      p.OptionPrix = participant.OptionPrix;
    }
  });
  const res = await Promise.all(
    others.map((p) => controller.ParticipantsUpdate(p))
  );
  if (res.includes(undefined)) return;
  controller.showMessage("Réglages des participants modifiés avec succès.");
  ensureDossier();
}

async function createAide(params: AidesCreateIn) {
  const res = await controller.AidesCreate(params);
  if (res === undefined) return;
  controller.showMessage("Aide ajoutée avec succès.");

  // la nouvelle aide ne modifie pas le statut du dossier :
  // pas de besoin de [loadDossier]
  if (dossierDetails.value == null) return;
  const allAides = dossierDetails.value.Dossier.Aides || {};
  const thisPart = allAides[params.IdParticipant] || {};
  thisPart[res.Id] = res;
  allAides[params.IdParticipant] = thisPart;
  dossierDetails.value.Dossier.Aides = allAides;
}

async function updateAide(aide: Aide) {
  const res = await controller.AidesUpdate(aide);
  if (res === undefined) return;
  controller.showMessage("Aide modifiée avec succès.");
  ensureDossier();
}

async function deleteAide(id: IdAide) {
  const res = await controller.AidesDelete({ id });
  if (res === undefined) return;
  controller.showMessage("Aide supprimée avec succès.");
  ensureDossier();
}

async function createPaiement() {
  if (dossierDetails.value == null) return;
  const res = await controller.PaiementsCreate({
    "id-dossier": dossierDetails.value.Dossier.Dossier.Id,
  });
  if (res === undefined) return;
  controller.showMessage("Paiement créé avec succès.");
  ensureDossier();
  // show edit dialog
  detailsPannel.value?.showEditPaiement(res);
}

async function updatePaiement(paiement: Paiement) {
  const res = await controller.PaiementsUpdate(paiement);
  if (res === undefined) return;
  controller.showMessage("Paiement modifié avec succès.");
  ensureDossier();
}
</script>
