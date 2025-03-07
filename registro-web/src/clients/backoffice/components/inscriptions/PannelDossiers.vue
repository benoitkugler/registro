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
          dossiersCount == null
            ? '-'
            : `${dossiersCount.length} / ${dossiersCount.total}`
        "
      >
        <template #append>
          <v-btn icon size="small" @click="showCreateDossier = true">
            <v-icon color="green">mdi-plus</v-icon>
          </v-btn>
        </template>

        <v-card-text>
          <DossierList
            :camps="allCamps"
            v-model:query="query"
            @click="onSelectDossier"
            @update="onListChange"
            ref="dossierList"
          ></DossierList>
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
        @merge-dossier="mergeDossier"
        @create-participant="createParticipant"
        @update-participant="updateParticipant"
        @delete-participant="deleteParticipant"
        @expand-participant="expandParticipant"
        @create-aide="createAide"
        @delete-aide="deleteAide"
        @update-aide="updateAide"
        @create-paiement="createPaiement"
        @update-paiement="updatePaiement"
        @delete-paiement="deletePaiement"
        @send-message="sendMessage"
        @delete-message="deleteMessage"
        ref="detailsPannel"
      ></DossierDetailsPannel>
      <div v-else class="text-center font-italic my-6">
        Sélectionner un dossier...
      </div>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { onMounted, ref, useTemplateRef } from "vue";
import {
  type CampItem,
  type SearchDossierIn,
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
  type SearchDossierOut,
  type DossiersMergeIn,
  type Event,
  type DossierHeader,
} from "../../logic/api";
import { controller, emptyQuery } from "../../logic/logic";
import DossierDetailsPannel from "./dossiers/DossierDetailsPannel.vue";
import CreateDossierCard from "./dossiers/CreateDossierCard.vue";
import DossierList from "./dossiers/DossierList.vue";

const props = defineProps<{}>();

defineExpose({ showDossier });

onMounted(() => {
  loadCamps();
  loadStructureaides();
});

const detailsPannel = useTemplateRef("detailsPannel");

const dossiersCount = ref<{ length: number; total: number } | null>(null);

const query = ref<SearchDossierIn>(emptyQuery());

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

function onListChange(v: SearchDossierOut) {
  dossiersCount.value = { length: v.Dossiers?.length || 0, total: v.Total };
  // if no dossier is yet selected, auto select the first
  if (dossierDetails.value == null && v.Dossiers?.length) {
    loadDossier(v.Dossiers[0].Id);
  }
}

const dossierList = useTemplateRef("dossierList");
function refreshDossierList() {
  dossierList.value?.searchDossiers();
}

// showDossier may be called by the parent when switching to this pannel;
//  it loads the dossier details
async function showDossier(id: IdDossier, responsable: string) {
  // reset the query so that the given Dossier is found
  const empty = emptyQuery();
  query.value = {
    IdCamp: empty.IdCamp,
    Attente: empty.Attente,
    Reglement: empty.Reglement,
    Pattern: responsable,
  };
  await loadDossier(id);
}

const dossierDetails = ref<DossierDetails | null>(null);
// fetch the complete information and displays it in the right pannel
async function loadDossier(id: IdDossier) {
  const res = await controller.DossiersLoad({ id: id });
  if (res === undefined) return;
  dossierDetails.value = res;
  detailsPannel.value?.scrollToLastEvent();
}

async function onSelectDossier(d: DossierHeader) {
  loadDossier(d.Id);
  const res = await controller.EventsMarkMessagesSeen({ idDossier: d.Id });
  if (res === undefined) return;
  // update the list
  refreshDossierList();
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
  // reset the search
  refreshDossierList();
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
  // reset the search
  refreshDossierList();
}

async function mergeDossier(args: DossiersMergeIn) {
  const res = await controller.DossiersMerge(args);
  if (res === undefined) return;

  controller.showMessage("Dossier fusionné avec succès.");
  // reset the search
  refreshDossierList();

  loadDossier(args.To);
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
    idDossier: dossierDetails.value.Dossier.Dossier.Id,
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

async function deletePaiement(paiement: Paiement) {
  const res = await controller.PaiementsDelete({ id: paiement.Id });
  if (res === undefined) return;
  controller.showMessage("Paiement supprimé avec succès.");
  ensureDossier();
}

async function sendMessage(contenu: string) {
  if (dossierDetails.value == null) return;
  const res = await controller.EventsSendMessage({
    IdDossier: dossierDetails.value.Dossier.Dossier.Id,
    Contenu: contenu,
  });
  if (res === undefined) return;
  controller.showMessage("Message envoyé avec succès.");
  ensureDossier();
}

async function deleteMessage(event: Event) {
  const res = await controller.EventsDelete({
    id: event.Id,
  });
  if (res === undefined) return;
  controller.showMessage("Message supprimé avec succès.");
  ensureDossier();
}
</script>
