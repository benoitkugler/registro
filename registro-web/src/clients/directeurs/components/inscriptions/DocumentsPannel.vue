<template>
  <v-card
    v-if="data != null"
    title="Documents"
    subtitle="Documents déposés via l'espace de suivi, à télécharger"
  >
    <template #append>
      <v-btn
        class="mx-1"
        @click="
          selectedRelance = withMissingDocuments();
          showRelanceConfirmation = true;
        "
        :disabled="isSendingMailFor.size != 0"
      >
        <template #prepend>
          <v-icon>mdi-email-arrow-right</v-icon>
        </template>
        Relancer...</v-btn
      >
      <v-btn>
        <template #prepend>
          <v-icon>mdi-download</v-icon>
        </template>
        Télécharger...
        <v-menu activator="parent">
          <v-list>
            <v-list-item
              title="Fiches sanitaires"
              subtitle="Télécharger un fichier .pdf"
              :href="
                controller.DocumentsDownloadFichesSanitaires(
                  controller.authToken
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
                  controller.authToken
                )
              "
            >
            </v-list-item>
          </v-list>
        </v-menu>
      </v-btn>
    </template>
    <v-card-text>
      <v-table>
        <tr>
          <th class="text-left">Participant</th>
          <th>Fiche sanitaire</th>
          <th v-for="demande in demandes">
            {{ demande.title }}
          </th>
          <!-- actions -->
          <th style="width: 40px"></th>
        </tr>
        <tr v-for="participant in data.Participants">
          <td class="text-left py-1">{{ participant.Personne }}</td>
          <td class="text-center">
            <v-chip
              :color="
                Formatters.colorFichesanitaireState(participant.Fichesanitaire)
              "
              density="compact"
            >
              {{ FichesanitaireStateLabels[participant.Fichesanitaire] }}
            </v-chip>
          </td>
          <td class="text-center" v-for="demande in demandes">
            <v-btn
              v-if="(participant.Files || {})[demande.Id]?.length"
              icon
              size="small"
              flat
            >
              <v-icon color="green">mdi-check</v-icon>
              <v-menu activator="parent">
                <v-list>
                  <v-list-item
                    v-for="file in (participant.Files || {})[demande.Id]"
                    :title="file.NomClient"
                    :subtitle="`${Formatters.size(
                      file.Taille
                    )} - Ajouté ${Formatters.date(file.Uploaded)}`"
                  >
                    <template #append>
                      <FileCardReadonly
                        class="ma-2"
                        :file-key="file.Key"
                        :is-generated-doc="false"
                      ></FileCardReadonly>
                    </template>
                  </v-list-item>
                </v-list>
              </v-menu>
            </v-btn>

            <v-icon v-else color="orange">mdi-alert </v-icon>
          </td>
          <td class="text-right py-1">
            <v-btn
              icon
              size="small"
              :disabled="
                isComplete(participant) || isSendingMailFor.has(participant.Id)
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
  </v-card>
  <v-skeleton-loader v-else type="card"></v-skeleton-loader>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed, reactive } from "vue";
import { controller } from "../../logic/logic";
import {
  Categorie,
  CategorieLabels,
  FichesanitaireState,
  FichesanitaireStateLabels,
  type IdParticipant,
  type ParticipantDocuments,
  type ParticipantsDocuments,
} from "../../logic/api";
import { Formatters } from "@/utils";
import FileCardReadonly from "@/components/files/FileCardReadonly.vue";

const props = defineProps<{}>();

onMounted(loadDocuments);

const data = ref<ParticipantsDocuments | null>(null);
async function loadDocuments() {
  const res = await controller.ParticipantsLoadFiles();
  if (res === undefined) return;
  data.value = res;
}

const demandes = computed(() =>
  Object.values(data.value?.Demandes || {})
    .sort((a, b) => a.Id - b.Id)
    .map((d) => ({
      Id: d.Id,
      title:
        d.Categorie == Categorie.NoBuiltin
          ? d.Description
          : CategorieLabels[d.Categorie],
    }))
);

function isComplete(participant: ParticipantDocuments) {
  const demandes = Object.values(data.value?.Demandes || {});
  if (!demandes) return false;
  return (
    participant.Fichesanitaire == FichesanitaireState.UpToDate &&
    demandes.every((d) => ((participant.Files || {})[d.Id] || []).length != 0)
  );
}

const isSendingMailFor = reactive(new Set<IdParticipant>());
async function relanceDocuments(participant: ParticipantDocuments) {
  isSendingMailFor.add(participant.Id);
  const res = await controller.ParticipantsRelanceDocuments([participant.Id]);
  isSendingMailFor.delete(participant.Id);
  if (res === undefined) return;
  controller.showMessage(
    `Message de relance pour ${participant.Personne} envoyé avec succès.`
  );
}

const showRelanceConfirmation = ref(false);
const selectedRelance = ref<IdParticipant[]>([]);

function withMissingDocuments() {
  return (data.value?.Participants || [])
    .filter((p) => !isComplete(p))
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
