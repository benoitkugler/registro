<template>
  <v-card title="Participants" subtitle="Choix des séjours">
    <template #append>
      <v-btn color="green" @click="addParticipant">
        <template #prepend>
          <v-icon>mdi-plus</v-icon>
        </template>
        Ajouter un participant</v-btn
      >
    </template>
    <v-card-text>
      <ParticipantRow
        v-for="(participant, i) in participants"
        :model-value="participant"
        @update:model-value="(v) => (participants[i] = v)"
        :camps="props.camps"
        @delete="deleteParticipant(i)"
      ></ParticipantRow>
    </v-card-text>
  </v-card>
  <div ref="bottom"></div>
</template>

<script lang="ts" setup>
import { nextTick, onMounted, useTemplateRef } from "vue";
import {
  Nationnalite,
  Sexe,
  type CampExt,
  type Date_,
  type IdCamp,
  type Participant,
  type Pays,
  type ResponsableLegal,
} from "../logic/api";
import ParticipantRow from "./ParticipantRow.vue";

const props = defineProps<{
  camps: CampExt[];
  responsable: ResponsableLegal;
  preselected: IdCamp;
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

const bottomRef = useTemplateRef("bottom");

function nationnaliteFromPays(s: Pays): Nationnalite {
  if (s == "FR") return Nationnalite.Francaise;
  if (s == "CH") return Nationnalite.Suisse;
  return Nationnalite.Autre;
}

// prend en compte une éventulle pré-sélection du séjour
// copie les données du responsable pour le premier participant
function emptyParticipant() {
  const out: Participant = {
    Nom: props.responsable.Nom,
    Prenom: "",
    DateNaissance: "0001-01-01" as Date_,
    Sexe: Sexe.NoSexe,
    Nationnalite: nationnaliteFromPays(props.responsable.Pays),
    IdCamp: props.preselected,
  };
  if (!participants.value.length) {
    out.Prenom = props.responsable.Prenom;
    out.DateNaissance = props.responsable.DateNaissance;
    out.Sexe = props.responsable.Sexe;
  }
  return out;
}

function addParticipant() {
  participants.value.push(emptyParticipant());
  nextTick(() => {
    if (bottomRef.value != null) bottomRef.value.scrollIntoView(true);
  });
}

function deleteParticipant(index: number) {
  participants.value.splice(index, 1);
}
</script>
