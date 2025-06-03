<template>
  <v-card v-if="data != null" title="Participants">
    <template #append>
      <v-btn @click="showDocuments = true" class="mx-2">
        <template #prepend>
          <v-icon>mdi-folder</v-icon>
        </template>
        Documents</v-btn
      >
      <v-btn @click="showFichesSanitaires = true">
        <template #prepend>
          <v-icon>mdi-pill</v-icon>
        </template>
        Fiches sanitaires</v-btn
      >
      <v-divider thickness="1" vertical class="mx-2"></v-divider>
      <v-tooltip text="Trier par moment d'inscription">
        <template #activator="{ props: tooltipProps }">
          <v-btn
            v-bind="tooltipProps"
            icon
            :variant="sortByTime ? 'tonal' : 'flat'"
            @click="sortByTime = !sortByTime"
          >
            <v-icon>mdi-sort-calendar-ascending</v-icon>
          </v-btn>
        </template>
      </v-tooltip>
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
} from "../../logic/api";
import { Participants } from "@/utils";
import FichesSanitairesPannel from "./FichesSanitairesPannel.vue";
import DocumentsPannel from "./DocumentsPannel.vue";

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
  const out = (data.value || []).map((p) => p);
  out.sort((a, b) => Participants.cmp(a, b, sortByTime.value));
  return out;
});

const data = ref<ParticipantExt[] | null>(null);
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
</script>
