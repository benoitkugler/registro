<template>
  <v-card
    title="Documents du séjour"
    subtitle="Choisir les documents à lire ou remplir par les familles."
    class="ma-2 mx-auto"
    max-width="800px"
  >
    <template #append v-if="data">
      <v-chip v-if="data.Ready" prepend-icon="mdi-lock-open" color="success">
        Documents visibles
      </v-chip>
      <v-btn
        v-else
        @click="showSendDialog = true"
        prepend-icon="mdi-lock-open"
        :disabled="!allDocumentsToShow.length"
      >
        Envoyer</v-btn
      >
    </template>
    <v-skeleton-loader v-if="data == null"></v-skeleton-loader>
    <v-card-text v-else>
      <v-alert type="info" closable>
        La <b>fiche sanitaire</b> et les <b>vaccins</b> sont automatiquement
        demandés pour les participants mineurs.
      </v-alert>
      <v-list>
        <!-- generated documents -->
        <v-list-subheader>Documents générés</v-list-subheader>
        <v-list-item
          title="Lettre aux familles"
          subtitle="Document officiel pour le premier contact"
        >
          <template #append>
            <v-btn
              icon="mdi-pencil"
              size="small"
              class="mr-4"
              @click="emit('goTo', 'lettre')"
            ></v-btn>
            <v-switch
              color="primary"
              hide-details
              :model-value="data.ToShow.LettreDirecteur"
              @update:model-value="(b) => updateToShow('LettreDirecteur', b!)"
            ></v-switch>
          </template>
        </v-list-item>
        <v-list-item title="Liste de vêtements">
          <template #append>
            <v-btn
              icon="mdi-pencil"
              size="small"
              class="mr-4"
              @click="emit('goTo', 'vetements')"
            ></v-btn>
            <v-switch
              color="primary"
              hide-details
              :model-value="data.ToShow.ListeVetements"
              @update:model-value="(b) => updateToShow('ListeVetements', b!)"
            ></v-switch>
          </template>
        </v-list-item>
        <v-list-item
          title="Charte"
          subtitle="Contrat moral à accepter pour les participants agés de 12 ans ou plus"
        >
          <template #append>
            <v-switch
              color="primary"
              hide-details
              :model-value="data.ToShow.CharteParticipant"
              @update:model-value="(b) => updateToShow('CharteParticipant', b!)"
            ></v-switch> </template
        ></v-list-item>
        <v-list-item
          title="Liste des participants"
          subtitle="Permet par exemple d'organiser un co-voiturage"
        >
          <template #append>
            <v-switch
              color="primary"
              hide-details
              :model-value="data.ToShow.ListeParticipants"
              @update:model-value="(b) => updateToShow('ListeParticipants', b!)"
            ></v-switch> </template
        ></v-list-item>

        <v-divider thickness="1" class="my-2"></v-divider>

        <!-- documents to download -->
        <v-row no-gutters>
          <v-col align-self="center">
            <v-list-subheader>Documents à lire </v-list-subheader>
          </v-col>
          <v-col align-self="center" cols="auto" class="mx-2">
            <v-btn size="small" @click="showUploadToDownload = true">
              <template #prepend>
                <v-icon color="green">mdi-plus</v-icon>
              </template>
              Ajouter
            </v-btn>
          </v-col>
        </v-row>
        <v-list-item v-if="!data.FilesToDownload?.length">
          <i>Aucun document.</i>
        </v-list-item>
        <v-list-item
          v-for="file in data.FilesToDownload"
          :title="file.NomClient"
        >
          <template #append>
            <FileCard :file="file" @delete="fileToDelete = file"></FileCard>
          </template>
        </v-list-item>

        <v-divider thickness="1" class="my-2"></v-divider>

        <!-- documents to upload -->
        <v-row no-gutters>
          <v-col align-self="center">
            <v-list-subheader
              >Documents à fournir (collectés sur l'espace de suivi)
            </v-list-subheader>
          </v-col>
          <v-col align-self="center" cols="auto" class="mx-2">
            <v-btn size="small" @click="showAddDemandeDialog = true">
              <template #prepend>
                <v-icon color="green">mdi-plus</v-icon>
              </template>
              Ajouter
            </v-btn>
          </v-col>
        </v-row>

        <v-list-item v-if="!data.CampDemandes?.length">
          <i>Aucun document.</i>
        </v-list-item>
        <v-list-item
          v-for="demande in data.CampDemandes"
          :title="Formatters.demande(demande.Demande)"
        >
          <template #append>
            <v-btn icon size="x-small" @click="unapplyDemande(demande.Demande)">
              <v-icon color="red">mdi-close</v-icon>
            </v-btn>
          </template>
        </v-list-item>
      </v-list>
    </v-card-text>

    <v-dialog v-model="showUploadToDownload" max-width="600px">
      <v-card
        title="Charger un document"
        subtitle="Ce document sera distribué sur l'espace personnel des familles."
      >
        <v-card-text>
          <FileInput
            label="Document à distribuer"
            @update="(f) => (fileToDownload = f)"
          ></FileInput>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="green"
            :disabled="fileToDownload == null"
            @click="addToDownload"
            >Charger</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- delete file -->
    <v-dialog
      :model-value="fileToDelete != null"
      @update:model-value="fileToDelete = null"
      max-width="400px"
    >
      <v-card title="Confirmer la suppression" v-if="fileToDelete">
        <v-card-text>
          Confirmez-vous la suppression du document
          <i> {{ fileToDelete.NomClient }} </i> ? <br /><br />

          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="deleteFile">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
  <!-- demande selector/editor -->
  <v-dialog v-model="showAddDemandeDialog" max-width="800px">
    <AddDemandeCard
      v-if="data"
      :availableDemandes="data.AvailableDemandes || []"
      :selectedDemandes="(data.CampDemandes || []).map((d) => d.Demande.Id)"
      @selected="applyDemande"
      @created="onCreateDemande"
      @updated="onUpdateDemande"
      @deleted="onDeleteDemande"
    ></AddDemandeCard>
  </v-dialog>

  <!-- envoie documents -->
  <v-dialog v-model="showSendDialog" max-width="800px">
    <v-card title="Confirmer l'envoi des documents">
      <v-card-text>
        Les documents suivants seront disponibles sur l'espace de suivi, et un
        mail de notification sera envoyé.

        <div class="mt-2 text-center">
          <v-chip v-for="doc in allDocumentsToShow" class="mx-1">
            {{ doc }}</v-chip
          >
        </div>
      </v-card-text>
      <v-card-actions>
        <v-btn @click="unlockDocuments" prepend-icon="mdi-lock-open">
          Débloquer sans envoi de mail
        </v-btn>
        <v-spacer></v-spacer>
        <v-btn
          @click="unlockAndSendDocuments"
          prepend-icon="mdi-email-arrow-right"
        >
          Débloquer et notifier</v-btn
        >
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- progress monitor -->
  <v-dialog :model-value="sendingProgress != null" persistent max-width="600px">
    <RequestProgressCard
      v-if="sendingProgress"
      title="Envoi du sondage en cours"
      :current="sendingProgress.Current"
      :total="sendingProgress.Total"
    ></RequestProgressCard>
  </v-dialog>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from "vue";
import { controller } from "../../logic/logic";
import type {
  Demande,
  DemandeDirecteur,
  DocumentsOut,
  DocumentsToShow,
  PublicFile,
  SendProgress,
} from "../../logic/api";
import type { DocumentsTab } from "../../plugins/router";
import FileInput from "@/components/files/FileInput.vue";
import AddDemandeCard from "./AddDemandeCard.vue";
import { Formatters, readJSONStream } from "@/utils";
import type { Int } from "@/urls";
import RequestProgressCard from "@/components/RequestProgressCard.vue";

const props = defineProps<{}>();

const emit = defineEmits<{
  (e: "goTo", tab: DocumentsTab): void;
}>();

onMounted(fetchData);

const data = ref<DocumentsOut | null>(null);
async function fetchData() {
  const res = await controller.DocumentsGet();
  if (res === undefined) return;
  data.value = res;
}

async function updateToShow(field: keyof DocumentsToShow, b: boolean) {
  if (data.value == null) return;
  data.value.ToShow[field] = b;
  const res = await controller.DocumentsUpdateToShow(data.value.ToShow);
  if (res === undefined) return;
  controller.showMessage("Modifications enregistrées avec succès.");
}

const showUploadToDownload = ref(false);
const fileToDownload = ref<File | null>(null);
async function addToDownload() {
  showUploadToDownload.value = false;
  if (data.value == null || fileToDownload.value == null) return;
  const res = await controller.DocumentsUploadToDownload(fileToDownload.value);
  if (res === undefined) return;
  controller.showMessage("Document ajouté avec succès.");
  data.value.FilesToDownload = (data.value.FilesToDownload || []).concat(res);
}

const fileToDelete = ref<PublicFile | null>(null);
async function deleteFile() {
  const toDelete = fileToDelete.value;
  fileToDelete.value = null;
  if (data.value == null || toDelete == null) return;
  const res = await controller.DocumentsDeleteToDownload({
    key: toDelete.Key,
  });
  if (res === undefined) return;
  data.value.FilesToDownload = (data.value.FilesToDownload || []).filter(
    (f) => f.Key != toDelete.Key
  );
  controller.showMessage("Document supprimé avec succès.");
}

const showAddDemandeDialog = ref(false);

async function onCreateDemande(d: DemandeDirecteur) {
  if (!data.value) return;
  data.value.AvailableDemandes = (data.value.AvailableDemandes || []).concat(d);
  // also add it to the camp
  const res = await controller.DocumentsApplyDemande({
    idDemande: d.Demande.Id,
  });
  if (res === undefined) return;
  controller.showMessage("Demande créée et activée pour ce séjour.");
  data.value.CampDemandes = (data.value?.CampDemandes || []).concat(res);
}

function onUpdateDemande(demande: DemandeDirecteur) {
  if (!data.value) return;
  const index1 = (data.value!.AvailableDemandes || []).findIndex(
    (dd) => dd.Demande.Id == demande.Demande.Id
  );
  data.value.AvailableDemandes![index1] = demande;
  const index2 = (data.value!.CampDemandes || []).findIndex(
    (dd) => dd.Demande.Id == demande.Demande.Id
  );
  data.value.CampDemandes![index2] = demande;
}

function onDeleteDemande(d: Demande) {
  if (!data.value) return;
  data.value.AvailableDemandes = (data.value.AvailableDemandes || []).filter(
    (dd) => dd.Demande.Id != d.Id
  );
  data.value.CampDemandes = (data.value.CampDemandes || []).filter(
    (dd) => dd.Demande.Id != d.Id
  );
}

async function applyDemande(demande: Demande) {
  if (!data.value) return;
  showAddDemandeDialog.value = false;
  const res = await controller.DocumentsApplyDemande({ idDemande: demande.Id });
  if (res === undefined) return;
  controller.showMessage("Demande bien activée pour ce séjour.");
  data.value.CampDemandes = (data.value?.CampDemandes || []).concat(res);
}

async function unapplyDemande(demande: Demande) {
  if (!data.value) return;
  const res = await controller.DocumentsUnapplyDemande({
    idDemande: demande.Id,
  });
  if (res === undefined) return;
  controller.showMessage("Demande retirée avec succès.");
  data.value.CampDemandes = (data.value?.CampDemandes || []).filter(
    (d) => d.Demande.Id != demande.Id
  );
}

const showSendDialog = ref(false);
const allDocumentsToShow = computed(() => {
  const d = data.value;
  if (!d) return [];

  const out: string[] = [];
  if (d.ToShow.LettreDirecteur) out.push("Lettre aux familles");
  if (d.ToShow.ListeVetements) out.push("Liste de vêtements");
  if (d.ToShow.CharteParticipant) out.push("Charte");
  if (d.ToShow.ListeParticipants) out.push("Liste des participants");
  d.FilesToDownload?.forEach((doc) => out.push(doc.NomClient));
  d.CampDemandes?.forEach((doc) => out.push(Formatters.demande(doc.Demande)));
  return out;
});

async function unlockDocuments() {
  showSendDialog.value = false;
  const res = await controller.DocumentsUnlock();
  if (res === undefined) return;
  controller.showMessage("Documents débloqués avec succès.");
  if (data.value) data.value.Ready = true;
}

const sendingProgress = ref<SendProgress | null>(null);
async function unlockAndSendDocuments() {
  // start with initial 0 progress
  sendingProgress.value = {
    Current: 0 as Int,
    Total: 10 as Int, // just a guess
  };
  const res = await controller.DocumentsUnlockAndSend();
  if (res === undefined) {
    showSendDialog.value = false;
    sendingProgress.value = null;
    return;
  }
  await readJSONStream(
    res,
    (v) => (sendingProgress.value = v),
    (err) => controller.onError("Envoi d'une notification", err)
  );
  sendingProgress.value = null;
  showSendDialog.value = false;
  if (data.value) data.value.Ready = true;
  controller.showMessage("Notifications envoyées avec succès.");
}
</script>
