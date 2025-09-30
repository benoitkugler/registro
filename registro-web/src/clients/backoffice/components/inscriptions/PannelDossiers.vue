<template>
  <v-row no-gutters class="ma-2">
    <!-- liste de recherche -->
    <v-col cols="6">
      <v-card
        title="Dossiers"
        :subtitle="
          dossiersCount == null
            ? '-'
            : `${dossiersCount.length} affiché(s) / ${dossiersCount.total}`
        "
      >
        <template #append>
          <v-btn icon size="small" @click="showCreateDossier = true">
            <v-icon color="green">mdi-plus</v-icon>
          </v-btn>
          <v-divider thickness="1" vertical class="mx-2"></v-divider>
          <v-btn size="small">
            Actions
            <v-menu activator="parent">
              <v-list>
                <v-list-item
                  title="Envoyer les documents"
                  subtitle="d'un séjour"
                  prepend-icon="mdi-mail"
                  @click="showSendDocuments = true"
                >
                </v-list-item>
                <v-divider></v-divider>
                <v-list-item
                  title="Envoyer une relance de paiement"
                  prepend-icon="mdi-invoice-send"
                  @click="showRelancePaiement = true"
                >
                </v-list-item>
              </v-list>
            </v-menu>
          </v-btn>
        </template>

        <v-card-text>
          <DossierList
            :camps="allCamps"
            :allow-fonds-soutien="controller.isFondsSoutien"
            :selected="dossierDetails?.Dossier.Dossier.Id"
            v-model:query="query"
            @click="onSelectDossier"
            @update="onListChange"
            ref="dossierList"
          ></DossierList>
        </v-card-text>
      </v-card>
    </v-col>
    <!-- pannel de détails -->
    <v-col cols="6">
      <DossierDetailsPannel
        v-if="dossierDetails != null"
        :dossier="dossierDetails"
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
        @delete-file-aide="deleteFileAide"
        @upload-file-aide="uploadFileAide"
        @create-paiement="createPaiement"
        @update-paiement="updatePaiement"
        @delete-paiement="deletePaiement"
        @send-message="sendMessage"
        @delete-message="deleteMessage"
        @send-facture="sendFacture"
        ref="detailsPannel"
      ></DossierDetailsPannel>
      <div v-else class="text-center font-italic my-6">
        Sélectionner un dossier...
      </div>
    </v-col>

    <!-- create dossier dialog -->
    <v-dialog v-model="showCreateDossier" max-width="400px">
      <CreateDossierCard @create="createDossier"></CreateDossierCard>
    </v-dialog>

    <!-- send documents -->
    <v-dialog v-model="showSendDocuments" max-width="1000px">
      <SendDocumentsCampCard
        :camps="allCamps"
        @send="sendDocumentsCamp"
      ></SendDocumentsCampCard>
    </v-dialog>

    <!-- preview relance paiements -->
    <v-dialog v-model="showRelancePaiement" max-width="1000px">
      <SendRelancePaiementCard
        :camps="allCamps"
        @send="sendRelancePaiement"
      ></SendRelancePaiementCard>
    </v-dialog>

    <!-- monitor documents -->
    <v-dialog
      :model-value="documentsCampProgress != null"
      max-width="400px"
      persistent
    >
      <RequestProgressCard
        v-if="documentsCampProgress"
        title="Envoi des documents en cours"
        :current="documentsCampProgress.Current"
        :total="documentsCampProgress.Total"
      ></RequestProgressCard>
    </v-dialog>

    <!-- monitor relances -->
    <v-dialog
      :model-value="relancePaiementProgress != null"
      max-width="400px"
      persistent
    >
      <RequestProgressCard
        v-if="relancePaiementProgress"
        title="Envoi des relances en cours"
        :current="relancePaiementProgress.Current"
        :total="relancePaiementProgress.Total"
      ></RequestProgressCard>
    </v-dialog>
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
  type IdAide,
  type IdCamp,
  type SendProgress,
} from "../../logic/api";
import { controller, emptyQuery, idQuery } from "../../logic/logic";
import DossierDetailsPannel from "./dossiers/DossierDetailsPannel.vue";
import CreateDossierCard from "./dossiers/CreateDossierCard.vue";
import DossierList from "./dossiers/DossierList.vue";
import { watch } from "vue";
import SendDocumentsCampCard from "./dossiers/SendDocumentsCampCard.vue";
import SendRelancePaiementCard from "./dossiers/SendRelancePaiementCard.vue";
import { readJSONStream } from "@/utils";
import RequestProgressCard from "../RequestProgressCard.vue";
import type { Int } from "@/urls";

const props = defineProps<{
  initialDossier?: IdDossier;
}>();

onMounted(() => {
  loadCamps();
  if (props.initialDossier !== undefined) {
    showDossier(props.initialDossier);
  }
});

watch(
  () => props.initialDossier,
  () => {
    if (props.initialDossier !== undefined) {
      showDossier(props.initialDossier);
    }
  }
);

const detailsPannel = useTemplateRef("detailsPannel");

const dossiersCount = ref<{ length: number; total: number } | null>(null);

const query = ref<SearchDossierIn>(emptyQuery());

const allCamps = ref<CampItem[]>([]);
async function loadCamps() {
  const res = await controller.GetCamps();
  if (res === undefined) return;
  allCamps.value = res || [];
}

function onListChange(v: SearchDossierOut) {
  dossiersCount.value = { length: v.Dossiers?.length || 0, total: v.Total };
}

const dossierList = useTemplateRef("dossierList");
function refreshDossierList() {
  dossierList.value?.searchDossiers();
}

// showDossier may be called by the parent when switching to this pannel;
//  it loads the dossier details
async function showDossier(id: IdDossier) {
  await loadDossier(id);
  // reset the query so that the given Dossier is found
  query.value = idQuery(id);
}

const dossierDetails = ref<DossierDetails | null>(null);
// fetch the complete information and displays it in the right pannel
async function loadDossier(id: IdDossier) {
  const res = await controller.DossiersLoad({ id });
  if (res === undefined) return;
  dossierDetails.value = res;
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
  await showDossier(res.Id);
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

async function deleteFileAide(id: IdAide) {
  const res = await controller.AidesJustificatifDelete({ idAide: id });
  if (res === undefined) return;
  controller.showMessage("Justificatif supprimé avec succès.");
  if (dossierDetails.value != null && dossierDetails.value.Dossier.AidesFiles) {
    delete dossierDetails.value.Dossier.AidesFiles[id];
  }
}

async function uploadFileAide(f: File, id: IdAide) {
  const res = await controller.AidesJustificatifUpload(f, { idAide: id });
  if (res === undefined) return;
  controller.showMessage("Justificatif téléversé avec succès.");
  if (dossierDetails.value != null) {
    const m = dossierDetails.value.Dossier.AidesFiles || {};
    m[id] = res;
    dossierDetails.value.Dossier.AidesFiles = m;
  }
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

async function sendFacture() {
  if (dossierDetails.value == null) return;
  const res = await controller.EventsSendFacture({
    idDossier: dossierDetails.value.Dossier.Dossier.Id,
  });
  if (res === undefined) return;
  controller.showMessage("Demande de règlement envoyé avec succès.");
  ensureDossier();
}

const showSendDocuments = ref(false);
const documentsCampProgress = ref<SendProgress | null>(null);
async function sendDocumentsCamp(idCamp: IdCamp, idDossiers: IdDossier[]) {
  // start with initial 0 progress
  documentsCampProgress.value = {
    Current: 0 as Int,
    Total: idDossiers.length as Int,
  };
  const res = await controller.EventsSendDocumentsCamp({
    IdCamp: idCamp,
    IdDossiers: idDossiers,
  });
  if (res === undefined) {
    documentsCampProgress.value = null;
    return;
  }
  await readJSONStream(
    res,
    (v) => (documentsCampProgress.value = v),
    (err) => controller.onError("Envoi des documents", err)
  );
  documentsCampProgress.value = null;
  ensureDossier();
  controller.showMessage("Documents du séjour envoyés avec succès.");
}

const showRelancePaiement = ref(false);
const relancePaiementProgress = ref<SendProgress | null>(null);
async function sendRelancePaiement(idCamp: IdCamp, idDossiers: IdDossier[]) {
  // start with initial 0 progress
  relancePaiementProgress.value = {
    Current: 0 as Int,
    Total: idDossiers.length as Int,
  };
  const res = await controller.EventsSendRelancePaiement({
    IdDossiers: idDossiers,
  });
  if (res === undefined) {
    relancePaiementProgress.value = null;
    return;
  }
  await readJSONStream(
    res,
    (v) => (relancePaiementProgress.value = v),
    (err) => controller.onError("Envoi de la relance", err)
  );
  relancePaiementProgress.value = null;
  ensureDossier();
  controller.showMessage("Toutes les relances ont été envoyées avec succès.");
}
</script>
