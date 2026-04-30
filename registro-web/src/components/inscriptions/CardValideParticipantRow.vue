<template>
  <v-row>
    <v-col cols="3" align-self="center">
      <v-list-item
        :title="Personnes.label(props.participant.Personne)"
        :subtitle="Camps.label(props.participant.Camp)"
      ></v-list-item>
    </v-col>
    <v-col cols="5" align-self="center" class="text-center">
      <v-chip
        v-if="props.statut.Hint != StatutParticipant.Inscrit"
        color="warning"
        prepend-icon="mdi-alert"
      >
        {{ formatStatutCauses(props.statut.Causes) }}
      </v-chip>
      <v-chip v-if="props.statut.AllowedValidation?.length == 1" size="small">
        Seul le centre peut accepter cette inscription.
      </v-chip>
    </v-col>
    <v-col align-self="center">
      <StatutParticipantField
        v-if="props.statut.AllowedValidation?.length"
        v-model="selected"
        hide-details
        :restrict-items="props.statut.AllowedValidation || []"
      ></StatutParticipantField>
      <v-chip
        v-else-if="
          props.participant.Participant.Statut != StatutParticipant.AStatuer
        "
        size="small"
      >
        Inscription déjà validée.
      </v-chip>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { Camps, Formatters, Personnes } from "@/utils";
import {
  StatutParticipant,
  type ParticipantCamp,
  type StatutCauses,
  type StatutExt,
} from "../../clients/backoffice/logic/api";
const props = defineProps<{
  participant: ParticipantCamp;
  statut: StatutExt;
}>();

const selected = defineModel<StatutParticipant>({ required: true });

function formatStatutCauses(c: StatutCauses) {
  if (!c.Age) {
    const out = c.CauseAge.Jeune
      ? "Trop jeune"
      : `Trop âgé${Formatters.accord(props.participant.Personne.Sexe)}`;
    return `${out} : ${c.CauseAge.Age} ans; écart ${c.CauseAge.EcartInDays} j.`;
  } else if (!c.EquilibreGF) {
    return "Equilibre G./F.";
  } else if (!c.Place) {
    return "Camp complet";
  } else {
    return "";
  }
}
</script>
