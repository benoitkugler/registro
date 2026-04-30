<template>
  <v-card
    title="Valider l'inscription"
    subtitle="Un mail de confirmation va être envoyé."
  >
    <v-card-text>
      <CardValideParticipantRow
        v-for="p in participants"
        :participant="p"
        :statut="(props.inscription.StatutHints || {})[p.Participant.Id]"
        v-model="inner[p.Participant.Id]"
      ></CardValideParticipantRow>
    </v-card-text>
    <v-card-actions>
      <v-btn @click="emit('valide', inner, false)" :disabled="!isValid"
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
  type InscriptionExt,
  StatutParticipant,
} from "../../clients/backoffice/logic/api";
import CardValideParticipantRow from "./CardValideParticipantRow.vue";

const props = defineProps<{
  inscription: InscriptionExt;
  idParticipants?: IdParticipant[]; // only edit these participants
}>();

const emit = defineEmits<{
  (e: "valide", params: Statuts, sendMail: boolean): void;
}>();

// restricted to user choice
const participants = computed(() =>
  (props.inscription.Participants || []).filter((p) =>
    props.idParticipants
      ? props.idParticipants.includes(p.Participant.Id)
      : true
  )
);

// start with server hints (if validable), restricted if needed to validable participants
const inner = ref<Statuts>(
  Object.fromEntries(
    participants.value
      .filter(
        (p) =>
          (props.inscription.StatutHints || {})[p.Participant.Id]
            .AllowedValidation?.length
      )
      .map((p) => {
        const serverHint = (props.inscription.StatutHints || {})[
          p.Participant.Id
        ].Hint;
        const allowed =
          (props.inscription.StatutHints || {})[p.Participant.Id]
            .AllowedValidation || [];
        // make sure the preselected value is in the allowed list,
        // which is non empty here
        const statut = allowed.includes(serverHint) ? serverHint : allowed[0];
        return [p.Participant.Id, statut];
      })
  ) satisfies Statuts
);

type Statuts = { [key in IdParticipant]: StatutParticipant };

const isValid = computed(() => Object.values(inner.value).length > 0);
</script>
