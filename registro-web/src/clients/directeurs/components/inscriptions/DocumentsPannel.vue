<template>
  <v-card
    v-if="data != null"
    title="Documents"
    subtitle="Documents déposés via l'espace de suivi, à télécharger"
  >
    <template #append>
      <v-row no-gutters>
        <v-col>
          <v-select
            v-model="reviewFilter"
            class="my-3"
            density="compact"
            hide-details
            label="Filtrer les documents"
            :items="reviewFilterItems"
            style="width: 200px"
          ></v-select>
        </v-col>

        <v-divider vertical thickness="1" class="mx-2"></v-divider>

        <v-col cols="auto" align-self="center">
          <v-btn
            class="mx-1"
            @click="
              selectedRelance = withMissingDocuments();
              showRelanceConfirmation = true;
            "
            :disabled="isSendingMailFor.size != 0"
            prepend-icon="mdi-email-arrow-right"
          >
            Relancer...</v-btn
          >
          <v-btn prepend-icon="mdi-download">
            Télécharger...
            <v-menu activator="parent">
              <v-list>
                <v-list-item
                  title="Fiches sanitaires"
                  subtitle="Télécharger un fichier .pdf"
                  :href="
                    controller.DocumentsDownloadFichesSanitaires(
                      controller.authToken,
                    )
                  "
                >
                </v-list-item>
                <v-list-item
                  v-for="demande in demandes"
                  :title="demande.title"
                  subtitle="Télécharger une archive .zip"
                  :href="
                    controller.DocumentsStreamFiles(
                      demande.Id,
                      controller.authToken,
                    )
                  "
                >
                </v-list-item>
              </v-list>
            </v-menu>
          </v-btn>
        </v-col>
      </v-row>
    </template>
    <v-card-text>
      <v-alert type="warning" v-if="!data.DocumentsReady">
        Les documents sont encore verrouillés et masqués aux participants.
        <template #append>
          <v-btn prepend-icon="mdi-cog" :to="{ path: '/documents' }">
            Configurer
          </v-btn>
        </template>
      </v-alert>

      <v-table class="mt-4">
        <tr>
          <th class="text-left">Participant</th>
          <th>Fiche sanitaire</th>
          <th v-for="demande in demandes">
            {{ demande.title }}
          </th>
          <!-- actions -->
          <th style="width: 40px"></th>
        </tr>
        <tr v-for="participant in filteredParticipants">
          <td class="text-left py-1">{{ participant.Personne }}</td>
          <!-- fiche sanitaire -->
          <td class="text-center">
            <span v-if="!participant.FicheSanitaireRequired">-</span>
            <v-btn
              v-else-if="participant.FichesanitaireValid"
              icon
              size="small"
              flat
            >
              <v-icon
                :icon="
                  Formatters.reviewStatus(
                    participant.Fichesanitaire.ReviewStatus,
                  ).icon
                "
                :color="
                  Formatters.reviewStatus(
                    participant.Fichesanitaire.ReviewStatus,
                  ).color
                "
              ></v-icon>
              <v-menu activator="parent">
                <v-card min-width="600px">
                  <v-card-text>
                    <v-row no-gutters>
                      <v-col cols="3" align-self="center">
                        <v-list-item-title>Fiche sanitaire</v-list-item-title>
                      </v-col>
                      <v-col>
                        <v-btn
                          @click="downloadFicheSanitaire(participant.Id)"
                          icon
                        >
                          <v-icon>mdi-file-pdf-box</v-icon>
                        </v-btn>
                      </v-col>

                      <v-spacer></v-spacer>

                      <DocumentReviewControls
                        :status="participant.Fichesanitaire.ReviewStatus"
                        :is-sending-review="false"
                        @accept="
                          updateFileReview(
                            participant.Id,
                            { Valid: false, Id: 0 as IdFile }, // fiche sanitaire mode
                            true,
                          )
                        "
                        @reject="prepareRejectionFichesanitaire(participant.Id)"
                      ></DocumentReviewControls>
                    </v-row>
                  </v-card-text>
                </v-card>
              </v-menu>
            </v-btn>
            <v-icon v-else color="orange">mdi-alert </v-icon>
          </td>

          <td class="text-center" v-for="demande in demandes">
            <span v-if="!(participant.DemandesRequired || {})[demande.Id]"
              >-</span
            >
            <v-btn
              v-else-if="filesFor(participant, demande.Id).length"
              icon
              size="small"
              flat
            >
              <v-icon
                :color="
                  Formatters.reviewStatus(
                    dominantStatus(filesFor(participant, demande.Id)),
                  ).color
                "
                :icon="
                  Formatters.reviewStatus(
                    dominantStatus(filesFor(participant, demande.Id)),
                  ).icon
                "
              ></v-icon>
              <v-menu activator="parent">
                <v-card min-width="700px">
                  <v-card-text>
                    <v-row
                      no-gutters
                      v-for="file in filesFor(participant, demande.Id)"
                    >
                      <v-col cols="6" align-self="center">
                        <v-list-item-title>{{
                          file.NomClient
                        }}</v-list-item-title>
                        <v-list-item-subtitle>{{
                          `${Formatters.size(
                            file.Taille,
                          )} - Ajouté ${Formatters.date(file.Uploaded)}`
                        }}</v-list-item-subtitle>
                      </v-col>
                      <v-col cols="auto">
                        <FileCardReadonly
                          class="ma-1"
                          :file-key="file.Key"
                          :is-generated-doc="false"
                        ></FileCardReadonly>
                      </v-col>
                      <v-spacer></v-spacer>
                      <DocumentReviewControls
                        :status="file.ReviewStatus"
                        :is-sending-review="isSendingReview.has(file.Id)"
                        @accept="
                          updateFileReview(
                            participant.Id,
                            { Id: file.Id, Valid: true },
                            true,
                          )
                        "
                        @reject="prepareRejection(participant.Id, file)"
                      ></DocumentReviewControls>
                    </v-row>
                  </v-card-text>
                </v-card>
              </v-menu>
            </v-btn>
            <v-icon v-else color="orange">mdi-alert </v-icon>
          </td>
          <td class="text-right py-1">
            <v-btn
              icon
              size="small"
              :disabled="
                participant.IsComplete || isSendingMailFor.has(participant.Id)
              "
              @click="relanceDocuments(participant)"
            >
              <v-progress-circular
                size="small"
                v-if="isSendingMailFor.has(participant.Id)"
              ></v-progress-circular>
              <v-icon v-else icon="mdi-email-arrow-right"></v-icon>
              <v-tooltip activator="parent">
                Envoyer un email de relance
              </v-tooltip>
            </v-btn>
          </td>
        </tr>
      </v-table>
    </v-card-text>

    <v-dialog v-model="showRelanceConfirmation" max-width="600px">
      <v-card
        title="Envoi d'une relance"
        subtitle="Un mail de relance va être envoyé aux participants suivants."
      >
        <v-card-text>
          <v-autocomplete
            density="compact"
            multiple
            chips
            closable-chips
            :items="
              (data.Participants || []).map((p) => ({
                title: p.Personne,
                value: p.Id,
              }))
            "
            v-model="selectedRelance"
          ></v-autocomplete>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            @click="relanceDocumentsMany"
            :disabled="!selectedRelance.length"
            color="green"
            >Relancer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog
      :model-value="rejection != null"
      @update:model-value="rejection = null"
      max-width="600px"
    >
      <v-card
        title="Refuser ce document"
        :subtitle="rejection.fileName"
        v-if="rejection"
      >
        <v-card-text>
          <v-textarea
            v-model="rejection.reason"
            label="Motif (facultatif)"
            rows="3"
            auto-grow
          ></v-textarea>
          <p>
            Le motif sera envoyé par mail au responsable et ne sera pas
            conservé.
          </p>
        </v-card-text>
        <v-card-actions>
          <v-btn @click="rejection = null">Annuler</v-btn>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="confirmRejection">
            Refuser et prévenir
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
  <v-skeleton-loader v-else type="card"></v-skeleton-loader>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed, reactive } from "vue";
import { controller } from "../../logic/logic";
import {
  Categorie,
  CategorieLabels,
  FileReviewStatus,
  FileReviewStatusLabels,
  FichesanitaireState,
  FichesanitaireStateLabels,
  type IdDemande,
  type IdParticipant,
  type PublicFile,
  type ParticipantFiles,
  type ParticipantsFiles,
  ReviewNotif,
  type FicheSanitaireExt,
  type OptID_IdFile,
  type IdFile,
} from "../../logic/api";
import { Formatters, saveBlobAsFile } from "@/utils";
import FileCardReadonly from "@/components/files/FileCardReadonly.vue";
import DocumentReviewControls from "./DocumentReviewControls.vue";
import type { Value } from "sass";

const props = defineProps<{}>();

onMounted(loadDocuments);

const data = ref<ParticipantsFiles | null>(null);
async function loadDocuments() {
  const res = await controller.ParticipantsLoadFiles();
  if (res === undefined) return;
  data.value = res;
}

// all by default
const reviewFilter = ref<FileReviewStatus | null>(null);
const reviewFilterItems = [
  { title: "En attente", value: FileReviewStatus.FileReviewPending },
  { title: "Refusés", value: FileReviewStatus.FileReviewRejected },
  { title: "Validés", value: FileReviewStatus.FileReviewValidated },
  { title: "Tous", value: null },
];

const filteredParticipants = computed(() =>
  (data.value?.Participants || []).filter((participant) => {
    const status = reviewFilter.value;
    if (status === null) return true;
    const files = Object.values(participant.Files || {}).flatMap(
      (demandeFiles) => demandeFiles || [],
    );
    return files.some((file) => file.ReviewStatus === status);
  }),
);

const demandes = computed(() =>
  // since Vaccins is the only builtin, it has a smaller Id
  // and will appear first
  Object.values(data.value?.Demandes || {})
    .sort((a, b) => a.Id - b.Id)
    .map((d) => ({
      Id: d.Id,
      Categorie: d.Categorie,
      title:
        d.Categorie == Categorie.NoBuiltin
          ? d.Description
          : CategorieLabels[d.Categorie],
    })),
);

function filesFor(participant: ParticipantFiles, idDemande: IdDemande) {
  return participant.Files?.[idDemande] || [];
}

function dominantStatus(files: PublicFile[]) {
  if (
    files.some(
      (file) => file.ReviewStatus === FileReviewStatus.FileReviewRejected,
    )
  )
    return FileReviewStatus.FileReviewRejected;
  if (
    files.some(
      (file) => file.ReviewStatus === FileReviewStatus.FileReviewPending,
    )
  )
    return FileReviewStatus.FileReviewPending;
  return FileReviewStatus.FileReviewValidated;
}

async function downloadFicheSanitaire(idParticipant: IdParticipant) {
  const res = await controller.ParticipantsDownloadFicheSanitaire({
    idParticipant,
  });
  if (res === undefined) return;
  controller.showMessage("Fiche sanitaire téléchargée avec succès.");
  saveBlobAsFile(res.blob, res.filename);
}

const isSendingReview = reactive(new Set<number>());
async function updateFileReview(
  idParticipant: IdParticipant,
  idFile: OptID_IdFile,
  accepted: boolean,
  reason = "",
) {
  if (idFile.Valid) isSendingReview.add(idFile.Id);
  const res = await controller.ParticipantsReviewFile({
    IdParticipant: idParticipant,
    IdFile: idFile,
    Accepted: accepted,
    Reason: reason,
  });
  if (idFile.Valid) isSendingReview.delete(idFile.Id);
  if (res === undefined) return;

  const index = data.value?.Participants?.findIndex(
    (p) => p.Id == idParticipant,
  )!;
  data.value!.Participants![index] = res.Files;

  controller.showMessage(
    res.Notif == ReviewNotif.RejectionMailSent
      ? "Document refusé ; le responsable a été prévenu."
      : res.Notif == ReviewNotif.CompletionMailSent
        ? "Document validé ; le responsable a été prévenu."
        : "Document validé",
  );
}

const rejection = ref<{
  idParticipant: IdParticipant;
  idFile: OptID_IdFile;
  fileName: string;
  reason: string;
} | null>(null);
function prepareRejection(idParticipant: IdParticipant, file: PublicFile) {
  rejection.value = {
    idParticipant,
    idFile: { Valid: true, Id: file.Id },
    fileName: file.NomClient,
    reason: "",
  };
}
function prepareRejectionFichesanitaire(idParticipant: IdParticipant) {
  rejection.value = {
    idParticipant,
    idFile: { Valid: false, Id: 0 as IdFile },
    fileName: "Fiche sanitaire",
    reason: "",
  };
}

async function confirmRejection() {
  if (rejection.value == null) return;
  await updateFileReview(
    rejection.value.idParticipant,
    rejection.value.idFile,
    false,
    rejection.value.reason,
  );
  rejection.value = null;
}

const isSendingMailFor = reactive(new Set<IdParticipant>());
async function relanceDocuments(participant: ParticipantFiles) {
  isSendingMailFor.add(participant.Id);
  const res = await controller.ParticipantsRelanceDocuments([participant.Id]);
  isSendingMailFor.delete(participant.Id);
  if (res === undefined) return;
  controller.showMessage(
    `Message de relance pour ${participant.Personne} envoyé avec succès.`,
  );
}

const showRelanceConfirmation = ref(false);
const selectedRelance = ref<IdParticipant[]>([]);

function withMissingDocuments() {
  return (data.value?.Participants || [])
    .filter((p) => !p.IsComplete)
    .map((p) => p.Id);
}

async function relanceDocumentsMany() {
  showRelanceConfirmation.value = false;
  const l = selectedRelance.value;
  l.forEach((id) => isSendingMailFor.add(id));
  const res = await controller.ParticipantsRelanceDocuments(l);
  l.forEach((id) => isSendingMailFor.delete(id));
  if (res === undefined) return;
  controller.showMessage(`Messages de relance envoyés avec succès.`);
}
</script>
