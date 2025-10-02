<template>
  <v-card
    title="Participants"
    subtitle="Merci de préciser les participants à inscrire."
  >
    <template #append v-if="!isStart">
      <v-btn color="green" @click="addParticipant">
        <template #prepend>
          <v-icon>mdi-plus</v-icon>
        </template>
        Ajouter un participant</v-btn
      >
    </template>
    <v-card-text v-if="isStart">
      <v-row class="my-4">
        <v-col class="text-center">
          <v-btn size="large" class="text-none" @click="startWith('own')">
            <template #append>
              <v-icon>mdi-account</v-icon>
            </template>
            Je m'inscris comme participant.
          </v-btn>
        </v-col>
        <v-col class="text-center">
          <v-btn size="large" class="text-none" @click="startWith('one')">
            <template #append>
              <v-icon>mdi-account-child</v-icon>
            </template>
            J'inscris un participant.
          </v-btn>
        </v-col>
        <v-col class="text-center">
          <v-btn size="large" class="text-none" @click="startWith('two')">
            <template #append>
              <v-icon>mdi-account-child</v-icon>
            </template>
            J'inscris deux participants (ou plus).
          </v-btn>
        </v-col>
      </v-row>
    </v-card-text>
    <v-card-text v-else>
      <ParticipantRow
        v-for="(participant, i) in participants"
        :camps="props.camps"
        :settings="props.settings"
        :model-value="participant"
        @update:model-value="(v) => (participants[i] = v)"
        @delete="deleteParticipant(i)"
      ></ParticipantRow>
    </v-card-text>
  </v-card>
  <div ref="bottom"></div>
</template>

<script lang="ts" setup>
import { nextTick, onMounted, useTemplateRef } from "vue";
import {
  Sexe,
  type CampExt,
  type Date_,
  type IdCamp,
  type Nationnalite,
  type Participant,
  type Pays,
  type ResponsableLegal,
  type Settings,
} from "../logic/api";
import ParticipantRow from "./ParticipantRow.vue";
import { ref } from "vue";

const props = defineProps<{
  camps: CampExt[];
  responsable: ResponsableLegal;
  preselected: IdCamp;
  settings: Settings;
}>();

const participants = defineModel<Participant[]>({ required: true });

onMounted(() =>
  nextTick(() => {
    // make sure at least one participant is defined
    if (!participants.length) {
      addParticipant();
    }
  })
);

const isStart = ref(participants.value.length == 0);

const bottomRef = useTemplateRef("bottom");

function nationnaliteFromPays(s: Pays): Nationnalite {
  if (s == "CH") return { IsSuisse: true };
  return { IsSuisse: false };
}

// prend en compte une éventulle pré-sélection du séjour,
// et copie Nom et Nationnalite
function newParticipant() {
  const out: Participant = {
    Nom: props.responsable.Nom,
    Prenom: "",
    DateNaissance: "0001-01-01" as Date_,
    Sexe: Sexe.NoSexe,
    Nationnalite: nationnaliteFromPays(props.responsable.Pays),
    IdCamp: props.preselected,
  };
  return out;
}

function startWith(mode: "own" | "one" | "two") {
  switch (mode) {
    case "own":
      const newP = newParticipant();
      newP.Prenom = props.responsable.Prenom;
      newP.DateNaissance = props.responsable.DateNaissance;
      newP.Sexe = props.responsable.Sexe;
      participants.value = [newP];
      break;
    case "one":
      participants.value = [newParticipant()];
      break;
    case "two":
      participants.value = [newParticipant(), newParticipant()];
      break;
  }
  isStart.value = false;
}

function addParticipant() {
  participants.value.push(newParticipant());
  nextTick(() => {
    if (bottomRef.value != null) bottomRef.value.scrollIntoView(true);
  });
}

function deleteParticipant(index: number) {
  participants.value.splice(index, 1);
}
</script>
