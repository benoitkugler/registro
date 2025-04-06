<template>
  <v-card
    title="Valider l'inscription"
    subtitle="Un mail de confirmation va être envoyé."
  >
    <v-card-text>
      <CardValideParticipantRow
        v-for="p in participants"
        :participant="p"
        :statut="props.statuts[p.Participant.Id]"
        v-model="inner[p.Participant.Id]"
      ></CardValideParticipantRow>
    </v-card-text>
    <v-card-actions>
      <!-- only for backoffice -->
      <v-btn
        @click="emit('valide', inner, false)"
        :disabled="!isValid"
        v-if="props.idCamp === undefined"
        >Valider sans notification</v-btn
      >
      <v-spacer></v-spacer>
      <v-btn @click="emit('valide', inner, true)" :disabled="!isValid"
        >Valider</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import {
  type IdParticipant,
  type Inscription,
  type IdCamp,
  type StatutExt,
  StatutParticipant,
} from "../../clients/backoffice/logic/api";
import CardValideParticipantRow from "./CardValideParticipantRow.vue";

const props = defineProps<{
  inscription: Inscription;
  statuts: { [key in IdParticipant]: StatutExt };
  idCamp?: IdCamp; // only edit these participants
}>();

const emit = defineEmits<{
  (e: "valide", params: Statuts, sendMail: boolean): void;
}>();

const participants = computed(() =>
  (props.inscription.Participants || []).filter((p) =>
    props.idCamp ? p.Camp.Id == props.idCamp : true
  )
);

// start with server hints, restricted if needed to participants
const inner = ref(
  Object.fromEntries(
    participants.value
      .filter((p) => props.statuts[p.Participant.Id].Validable)
      .map((p) => [p.Participant.Id, props.statuts[p.Participant.Id].Statut])
  ) as Statuts
);

type Statuts = { [key in IdParticipant]: StatutParticipant };

const isValid = computed(() => Object.values(inner.value).length > 0);
</script>
