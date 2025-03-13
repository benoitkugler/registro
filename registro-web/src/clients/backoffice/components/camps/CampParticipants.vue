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
            IdDossier: 0 as Int,
            IdPersonne: 0 as Int,
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
        <div class="text-center font-italic" v-if="!participants.length">
          Aucun participant n'est encore inscrit sur ce séjour.
        </div>
        <v-row v-for="p in participants" no-gutters class="mx-2">
          <v-col align-self="center" cols="3">
            <v-list-item :title="Personnes.NOMPrenom(p.Personne)">
              <template #prepend>
                <v-tooltip
                  v-if="p.Participant.Statut != ListeAttente.Inscrit"
                  :text="ListeAttenteLabels[p.Participant.Statut]"
                >
                  <template #activator="{ props }">
                    <v-icon v-bind="props"> mdi-clock </v-icon>
                  </template>
                </v-tooltip>

                <v-tooltip
                  v-else-if="p.HasBirthday"
                  :text="`${p.Personne.Prenom} a son anniveraire pendant le camp !`"
                >
                  <template #activator="{ props }">
                    <v-icon v-bind="props" color="amber"
                      >mdi-cake-variant</v-icon
                    >
                  </template>
                </v-tooltip>
              </template>
            </v-list-item>
          </v-col>

          <v-col align-self="center" cols="auto">
            <v-icon class="ma-2">
              {{ Formatters.sexeIcon(p.Personne.Sexe) }}
            </v-icon>
          </v-col>
          <v-col align-self="center" cols="auto">
            {{ Formatters.dateNaissance(p.Personne.DateNaissance) }}
          </v-col>
          <v-col align-self="center" cols="1" class="text-center">
            {{
              ageFrom(
                p.Personne.DateNaissance,
                new Date(data.Camp.Camp.DateDebut)
              )
            }}
            ans
          </v-col>
          <v-col align-self="center" class="text-center">
            {{
              p.Participant.Navette == Navette.NoBus
                ? "-"
                : NavetteLabels[p.Participant.Navette]
            }}
          </v-col>
          <v-col align-self="center">
            {{ p.Participant.Details }}
          </v-col>
          <v-spacer></v-spacer>
          <v-col cols="auto">
            <v-menu>
              <template #activator="{ props: menuProps }">
                <v-btn
                  v-bind="menuProps"
                  icon="mdi-dots-vertical"
                  size="x-small"
                ></v-btn>
              </template>
              <v-list density="comfortable">
                <v-list-item
                  title="Modifier"
                  prepend-icon="mdi-pencil"
                  @click="toEdit = p"
                ></v-list-item>
                <v-divider></v-divider>
                <v-list-item
                  title="Aller à la personne"
                  @click="goToPersonne(p.Personne.Id)"
                ></v-list-item>
                <v-list-item
                  title="Aller au dossier"
                  @click="goToDossier(p.Participant.IdDossier)"
                ></v-list-item>
                <v-divider></v-divider>
                <v-list-item
                  title="Supprimer"
                  prepend-icon="mdi-delete"
                  @click="toDelete = p"
                ></v-list-item>
              </v-list>
            </v-menu>
          </v-col>
        </v-row>
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
  </v-card>
  <v-skeleton-loader v-else type="card"></v-skeleton-loader>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed } from "vue";
import { controller } from "@/clients/backoffice/logic/logic";
import {
  ListeAttente,
  ListeAttenteLabels,
  Navette,
  NavetteLabels,
  type CampItem,
  type CampsLoadOut,
  type IdCamp,
  type IdDossier,
  type IdPersonne,
  type Int,
  type Participant,
  type ParticipantPersonne,
  type ParticipantsCreateIn,
} from "@/clients/backoffice/logic/api";
import { Camps, Formatters, Personnes } from "@/utils";
import { ageFrom } from "@/components/date";
import ParticipantEdit from "./ParticipantEdit.vue";
import { goToDossier, goToPersonne } from "../../router";

const props = defineProps<{
  id: IdCamp;
}>();

const emit = defineEmits<{
  (e: "goToPersonne", id: IdPersonne): void;
  (e: "goToDossier", id: IdDossier): void;
}>();

onMounted(() => {
  loadCamp();
  fetchCamps();
});

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
  out.sort((a, b) => {
    const sa = a.Participant.Statut;
    const sb = b.Participant.Statut;
    // By liste attente : Inscrit is higher
    if (sa != sb) return sb - sa;
    // By name :
    return Personnes.NOMPrenom(a.Personne).localeCompare(
      Personnes.NOMPrenom(b.Personne)
    );
  });
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

const toEdit = ref<ParticipantPersonne | null>(null);
async function updateParticipant(p: Participant) {
  if (toEdit.value == null || data.value == null) return;
  const res = await controller.ParticipantsUpdate(p);
  toEdit.value = null;
  if (res === undefined) return;
  controller.showMessage("Participant modifié avec succès.");
  // if the personne has changed, reload the list
  loadCamp();
}

const toDelete = ref<ParticipantPersonne | null>(null);
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
</script>
