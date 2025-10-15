<template>
  <v-card
    v-if="data != null"
    title="Documents"
    subtitle="Documents déposés via l'espace de suivi, à télécharger"
  >
    <template #append>
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
      <v-table striped="even">
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
            <v-icon
              v-if="(participant.Files || {})[demande.Id]?.length"
              color="green"
              >mdi-check</v-icon
            >
            <v-icon v-else color="orange">mdi-alert</v-icon>
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
  </v-card>
  <v-skeleton-loader v-else type="card"></v-skeleton-loader>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed, reactive } from "vue";
import { controller, endpoints } from "../../logic/logic";
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
  const res = await controller.ParticipantsRelanceDocuments({
    idParticipant: participant.Id,
  });
  isSendingMailFor.delete(participant.Id);
  if (res === undefined) return;
  controller.showMessage(
    `Message de relance pour ${participant.Personne} envoyé avec succès.`
  );
}
</script>
