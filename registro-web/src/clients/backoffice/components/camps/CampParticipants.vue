<template>
  <v-card
    v-if="data != null"
    title="Participants"
    :subtitle="Camps.label(data.Camp.Camp)"
  >
    <template #append>
      <v-btn
        color="green"
        @click="
          toCreate = {
            IdCamp: props.id,
            IdDossier: 0 as IdDossier,
            IdPersonne: 0 as IdPersonne,
          }
        "
      >
        <template #prepend>
          <v-icon>mdi-plus</v-icon>
        </template>
        Créer un participant</v-btn
      >
    </template>
    <v-card-text class="mt-4">
      <v-skeleton-loader type="table" v-if="isLoading"></v-skeleton-loader>
      <div v-else>
        <ParticipantRowHeader></ParticipantRowHeader>

        <div class="text-center font-italic" v-if="!participants.length">
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
              <v-divider></v-divider>
              <v-list-item
                title="Changer de camp..."
                prepend-icon="mdi-file-move"
                @click="
                  moveArgs = {
                    Id: p.Participant.Id,
                    Target: 0 as IdCamp,
                  }
                "
              ></v-list-item>
              <v-list-item
                title="Notifier d'une place disponible."
                subtitle="Envoi un email"
                prepend-icon="mdi-email-alert"
                :disabled="
                  p.Participant.Statut == StatutParticipant.Inscrit ||
                  p.Participant.Statut == StatutParticipant.EnAttenteReponse
                "
                @click="confirmeSetPlaceLiberee = p"
              ></v-list-item>
              <v-divider></v-divider>
              <v-list-item
                title="Aller à la personne"
                :subtitle="`(ID : ${p.Personne.Id})`"
                @click="goToPersonne(p.Personne.Id)"
              ></v-list-item>
              <v-list-item
                title="Aller au dossier"
                :subtitle="`(ID : ${p.Participant.IdDossier})`"
                @click="goToDossier(p.Participant.IdDossier)"
              ></v-list-item>
              <v-divider></v-divider>
              <v-list-item
                title="Supprimer"
                prepend-icon="mdi-delete"
                @click="toDelete = p"
              ></v-list-item>
            </v-list>
          </template>
        </ParticipantRow>
      </div>
    </v-card-text>

    <!-- create participant -->
    <v-dialog
      v-if="toCreate != null"
      :model-value="toCreate != null"
      @update:model-value="toCreate = null"
      max-width="600px"
    >
      <v-card title="Créer un participant">
        <v-card-text>
          <v-row>
            <v-col cols="12">
              <SelectPersonne
                label="Participant"
                v-model="toCreate.IdPersonne"
                initial-personne=""
                :api="{
                  SelectPersonne: controller.SelectPersonne.bind(controller),
                }"
              ></SelectPersonne>
            </v-col>
            <v-col cols="12">
              <SelectDossier v-model="toCreate.IdDossier"></SelectDossier>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="green"
            @click="createParticipant"
            :disabled="!(toCreate.IdDossier && toCreate.IdPersonne)"
            >Créer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

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
        @save="updateParticipant"
      ></ParticipantEdit>
    </v-dialog>

    <!-- delete participant -->
    <v-dialog
      v-if="toDelete != null"
      :model-value="toDelete != null"
      @update:model-value="toDelete = null"
      max-width="600px"
    >
      <v-card title="Supprimer le participant">
        <v-card-text>
          Confirmez-vous la suppression du participant
          <b>{{ Personnes.label(toDelete.Personne) }}</b> ? <br />
          <br />

          Attention, cette opération est irréversible.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" @click="deleteParticipant">Supprimer</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- move participant -->
    <v-dialog
      v-if="moveArgs != null"
      :model-value="moveArgs != null"
      @update:model-value="moveArgs = null"
      max-width="500px"
    >
      <v-card title="Déplacer le participant">
        <v-card-text>
          <SelectCamp
            label="Nouveau séjour"
            :camps="camps"
            v-model="moveArgs.Target"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            @click="moveParticipant"
            :disabled="moveArgs.Target == props.id || !moveArgs.Target"
            >Déplacer</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- place libérée -->
    <v-dialog
      v-if="confirmeSetPlaceLiberee != null"
      :model-value="confirmeSetPlaceLiberee != null"
      @update:model-value="confirmeSetPlaceLiberee = null"
      max-width="500px"
    >
      <v-card title="Notifier d'une place disponible">
        <v-card-text>
          Confirmez-vous l'envoi d'une notification au responsable de
          {{ Personnes.label(confirmeSetPlaceLiberee.Personne) }} indiquant
          qu'une place est disponible ? <br /><br />

          Le nouveau statut de
          {{ confirmeSetPlaceLiberee.Personne.Prenom }} sera :
          <i>{{
            StatutParticipantLabels[StatutParticipant.EnAttenteReponse]
          }}</i
          >.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="setPlaceLiberee">Notifier</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
  <v-skeleton-loader v-else type="card"></v-skeleton-loader>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed, watch } from "vue";
import { controller } from "@/clients/backoffice/logic/logic";
import {
  StatutParticipant,
  StatutParticipantLabels,
  type CampItem,
  type CampsLoadOut,
  type IdCamp,
  type IdDossier,
  type IdPersonne,
  type Participant,
  type ParticipantExt,
  type ParticipantsCreateIn,
  type ParticipantsMoveIn,
} from "@/clients/backoffice/logic/api";
import { Camps, Participants, Personnes } from "@/utils";
import {
  goToDossier,
  goToParticipant,
  goToPersonne,
} from "../../plugins/router";

const props = defineProps<{
  id: IdCamp;
}>();

onMounted(() => {
  loadCamp();
  fetchCamps();
});

watch(() => props.id, loadCamp);

const isLoading = ref(false);

const camps = ref<CampItem[]>([]);
async function fetchCamps() {
  const res = await controller.GetCamps();
  if (res === undefined) return;
  camps.value = res || [];
}

// with sort
const participants = computed(() => {
  const out = (data.value?.Participants || []).map((p) => p);
  out.sort(Participants.cmp);
  return out;
});

const data = ref<CampsLoadOut | null>(null);
async function loadCamp() {
  isLoading.value = true;
  const res = await controller.CampsLoad({ idCamp: props.id });
  isLoading.value = false;
  if (res === undefined) return;
  data.value = res;
}

const toCreate = ref<ParticipantsCreateIn | null>(null);
async function createParticipant() {
  if (toCreate.value == null || data.value == null) return;
  const res = await controller.ParticipantsCreate(toCreate.value);
  toCreate.value = null;
  if (res === undefined) return;
  controller.showMessage("Participant créé avec succès.");
  data.value.Participants = (data.value.Participants || []).concat(res);
}

const toEdit = ref<ParticipantExt | null>(null);
async function updateParticipant(p: Participant) {
  if (toEdit.value == null || data.value == null) return;
  const res = await controller.ParticipantsUpdate(p);
  toEdit.value = null;
  if (res === undefined) return;
  controller.showMessage("Participant modifié avec succès.");
  // if the personne has changed, reload the list
  loadCamp();
}

const toDelete = ref<ParticipantExt | null>(null);
async function deleteParticipant() {
  if (toDelete.value == null || data.value == null) return;
  const id = toDelete.value.Participant.Id;
  const res = await controller.ParticipantsDelete({ id });
  toDelete.value = null;
  if (res === undefined) return;
  controller.showMessage("Participant supprimé avec succès.");
  data.value.Participants = (data.value.Participants || []).filter(
    (p) => p.Participant.Id != id
  );
}

const moveArgs = ref<ParticipantsMoveIn | null>(null);
async function moveParticipant() {
  const args = moveArgs.value;
  if (args == null || data.value == null) return;
  moveArgs.value = null;
  const res = await controller.ParticipantsMove(args);
  if (res === undefined) return;
  controller.showMessage("Participant déplacé avec succès.", undefined, {
    title: "Aller au nouveau séjour",
    action: () => goToParticipant({ IdCamp: args.Target, Id: args.Id }),
  });
  data.value.Participants = (data.value.Participants || []).filter(
    (p) => p.Participant.Id != args.Id
  );
}

const confirmeSetPlaceLiberee = ref<ParticipantExt | null>(null);
async function setPlaceLiberee() {
  if (confirmeSetPlaceLiberee.value == null) return;
  const id = confirmeSetPlaceLiberee.value.Participant.Id;
  confirmeSetPlaceLiberee.value = null;
  const res = await controller.ParticipantsSetPlaceLiberee({ id });
  if (res === undefined) return;
  // update the participant
  if (data.value == null) return;
  const item = data.value.Participants?.find((p) => p.Participant.Id == id)!;
  item.Participant = res;
}
</script>
