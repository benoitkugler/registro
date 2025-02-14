<template>
  <v-card title="Participants" subtitle="Choix des séjours">
    <template v-slot:append>
      <v-btn color="green" @click="addParticipant">
        <template v-slot:prepend>
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
</template>

<script lang="ts" setup>
import { ref } from "vue";
import {
  Nationnalite,
  Sexe,
  type CampExt,
  type Date_,
  type IdCamp,
  type Int,
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

function nationnaliteFromPays(s: Pays): Nationnalite {
  if (s == "FR") return Nationnalite.Francaise;
  if (s == "CH") return Nationnalite.Suisse;
  return Nationnalite.Autre;
}

// prend en compte une éventulle pré-sélection du séjour
function emptyParticipant(): Participant {
  return {
    Nom: props.responsable.Nom,
    Prenom: "",
    DateNaissance: "0001-01-01" as Date_,
    Sexe: Sexe.Empty,
    Nationnalite: nationnaliteFromPays(props.responsable.Pays),
    PreIdent: "",
    IdCamp: props.preselected,
  };
}

function addParticipant() {
  participants.value.push(emptyParticipant());
}

function deleteParticipant(index: number) {
  participants.value.splice(index, 1);
}
</script>
