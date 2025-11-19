<template>
  <v-card
    title="Participants"
    subtitle="Merci de préciser les participants à inscrire."
  >
    <template #append v-if="!isStart">
      <v-btn color="green" @click="addParticipant" prepend-icon="mdi-plus">
        Ajouter un participant</v-btn
      >
    </template>
    <v-card-text v-if="isStart">
      <v-row class="my-4" no-gutters justify="space-evenly">
        <v-col align-self="center" class="text-center" cols="12" md="5">
          <v-btn
            :size="smAndDown ? 'small' : 'large'"
            block
            class="text-none"
            @click="startWith('own')"
            prepend-icon="mdi-account"
          >
            Je m'inscris comme participant.
          </v-btn>
        </v-col>
        <v-col align-self="center" cols="12" md="auto">
          <v-divider :vertical="!smAndDown" thickness="4" class="my-4"
            >ou</v-divider
          >
        </v-col>
        <v-col align-self="center" class="text-center" cols="12" md="5">
          <v-btn
            :size="smAndDown ? 'small' : 'large'"
            block
            class="text-none"
            @click="startWith('one')"
            append-icon="mdi-account-child"
          >
            J'inscris un participant.
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
import { nextTick, useTemplateRef } from "vue";
import {
  Sexe,
  type CampExt,
  type Date_,
  type IdCamp,
  type Nationnalite,
  type Participant,
  type Pays,
  type ResponsableLegal,
  type ConfigInscription,
} from "../logic/api";
import ParticipantRow from "./ParticipantRow.vue";
import { ref } from "vue";
import { useDisplay } from "vuetify";

const props = defineProps<{
  camps: CampExt[];
  responsable: ResponsableLegal;
  preselected: IdCamp;
  settings: ConfigInscription;
}>();

const { smAndDown } = useDisplay();

const participants = defineModel<Participant[]>({ required: true });

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

function startWith(mode: "own" | "one") {
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
