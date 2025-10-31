<template>
  <v-card v-if="data != null">
    <template #title>
      <v-btn flat icon>
        <v-icon>mdi-information</v-icon>
        <v-tooltip activator="parent" content-class="ma-0 pa-0">
          <StatistiquesCard
            :participants="data.Participants || []"
          ></StatistiquesCard>
        </v-tooltip>
      </v-btn>
      Participants
    </template>
    <template #append>
      <v-btn @click="showMessages = true" class="mx-1">
        <template #prepend>
          <v-icon>mdi-email</v-icon>
        </template>
        Messages</v-btn
      >
      <v-btn @click="showDocuments = true" class="mx-1">
        <template #prepend>
          <v-icon>mdi-folder</v-icon>
        </template>
        Documents</v-btn
      >

      <v-divider thickness="1" vertical class="mx-1"></v-divider>

      <v-btn icon flat>
        <v-icon>mdi-dots-vertical</v-icon>
        <v-menu activator="parent">
          <v-list density="compact">
            <v-list-item
              title="Trier par moment d'inscription"
              @click="sortByTime = !sortByTime"
              prepend-icon="mdi-sort-calendar-ascending"
              :append-icon="sortByTime ? 'mdi-check' : ''"
            >
            </v-list-item>

            <v-divider thickness="1"></v-divider>

            <v-list-item
              prepend-icon="mdi-file-excel"
              title="Exporter"
              subtitle="au format Excel"
              link
              :href="endpoints.ParticipantsDownloadListe(controller.authToken)"
            ></v-list-item>
            <v-divider thickness="1"></v-divider>

            <v-list-item
              prepend-icon="mdi-pill"
              title="Fiches sanitaires"
              subtitle="Analyser le contenu des fiches sanitaires"
              @click="showFichesSanitaires = true"
            ></v-list-item>
            <v-list-item
              prepend-icon="mdi-currency-eur"
              title="Suivi du règlement"
              @click="showReglements = true"
            ></v-list-item>
          </v-list>
        </v-menu>
      </v-btn>
    </template>
    <v-card-text class="mt-4">
      <v-skeleton-loader type="table" v-if="isLoading"></v-skeleton-loader>
      <div v-else>
        <ParticipantRowHeader></ParticipantRowHeader>

        <div class="text-center font-italic mt-4" v-if="!participants.length">
          Aucun participant n'est encore inscrit sur ce séjour.
        </div>

        <ParticipantRow
          v-for="(p, index) in participants"
          :participant="p"
          :index="index"
        >
          <template #actions>
            <v-list density="comfortable">
              <v-list-item
                title="Modifier"
                prepend-icon="mdi-pencil"
                @click="toEdit = p"
              ></v-list-item>
            </v-list>
          </template>
        </ParticipantRow>
      </div>
    </v-card-text>

    <!-- edit participant -->
    <v-dialog
      v-if="toEdit != null"
      :model-value="toEdit != null"
      @update:model-value="toEdit = null"
      max-width="800px"
    >
      <ParticipantEdit
        :participant="toEdit.Participant"
        :personne="toEdit.Personne"
        :api="{
          SelectPersonne: controller.SelectPersonne.bind(controller),
        }"
        hide-personne-dossier
        readonly-statut
        @save="updateParticipant"
      ></ParticipantEdit>
    </v-dialog>

    <!-- fiches sanitaires -->
    <v-dialog v-model="showFichesSanitaires">
      <FichesSanitairesPannel></FichesSanitairesPannel>
    </v-dialog>

    <!-- documents demandés -->
    <v-dialog v-model="showDocuments" max-width="800px">
      <DocumentsPannel></DocumentsPannel>
    </v-dialog>

    <!-- messages -->
    <v-dialog v-model="showMessages" max-width="1200px">
      <MessagesPannel></MessagesPannel>
    </v-dialog>

    <!-- règlement -->
    <v-dialog v-model="showReglements" max-width="700px">
      <ReglementCard :data="data"></ReglementCard>
    </v-dialog>
  </v-card>
  <v-skeleton-loader v-else type="card"></v-skeleton-loader>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed } from "vue";
import { controller } from "../../logic/logic";
import {
  type CampItem,
  type Participant,
  type ParticipantExt,
  type ParticipantsOut,
} from "../../logic/api";
import { endpoints, Participants } from "@/utils";
import FichesSanitairesPannel from "./FichesSanitairesPannel.vue";
import DocumentsPannel from "./DocumentsPannel.vue";
import MessagesPannel from "./MessagesPannel.vue";
import StatistiquesCard from "./StatistiquesCard.vue";
import ReglementCard from "./ReglementCard.vue";

const props = defineProps<{}>();

onMounted(() => {
  loadParticipants();
  fetchCamps();
});

defineExpose({ loadParticipants });

const isLoading = ref(false);

const camps = ref<CampItem[]>([]);
async function fetchCamps() {
  const res = await controller.GetCamps();
  if (res === undefined) return;
  camps.value = res || [];
}

// with sort
const sortByTime = ref(false);
const participants = computed(() => {
  const out = (data.value?.Participants || []).map((p) => p);
  out.sort((a, b) => Participants.cmp(a, b, sortByTime.value));
  return out;
});

const data = ref<ParticipantsOut | null>(null);
async function loadParticipants() {
  isLoading.value = true;
  const res = await controller.ParticipantsGet();
  isLoading.value = false;
  if (res === undefined) return;
  data.value = res;
}

const toEdit = ref<ParticipantExt | null>(null);
async function updateParticipant(p: Participant) {
  if (toEdit.value == null || data.value == null) return;
  const res = await controller.ParticipantsUpdate(p);
  toEdit.value = null;
  if (res === undefined) return;
  controller.showMessage("Participant modifié avec succès.");
  // reload the list
  loadParticipants();
}

const showFichesSanitaires = ref(false);

const showDocuments = ref(false);

const showMessages = ref(false);

const showReglements = ref(false);
</script>
