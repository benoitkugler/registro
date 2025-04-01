<template>
  <v-row>
    <v-col cols="4" align-self="center">
      <v-list-item
        :title="Personnes.label(props.participant.Personne)"
        :subtitle="Camps.label(props.participant.Camp)"
      ></v-list-item>
    </v-col>
    <v-col cols="3" align-self="center" class="text-center">
      <v-chip
        v-if="props.statut.Statut != StatutParticipant.Inscrit"
        color="warning"
        prepend-icon="mdi-alert"
      >
        {{ formatStatutCauses(props.statut.Causes) }}
      </v-chip>
    </v-col>
    <v-col align-self="center">
      <StatutParticipantField
        v-if="props.statut.Validable"
        v-model="statut"
        hide-details
        :readonly="!props.statut.AllowedChanges?.length"
        :restrict-items="
          (props.statut.AllowedChanges || []).concat(props.statut.Statut)
        "
      ></StatutParticipantField>
      <v-chip v-else size="small">
        Seul le centre peut accepter cette inscription.
      </v-chip>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { Camps, Personnes } from "@/utils";
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

const statut = defineModel<StatutParticipant>({ required: true });

function formatStatutCauses(c: StatutCauses) {
  if (!c.AgeMin) {
    return "Trop jeune";
  } else if (!c.AgeMax) {
    return "Trop âgé";
  } else if (!c.EquilibreGF) {
    return "Equilibre G./F.";
  } else if (!c.Place) {
    return "Camp complet";
  } else {
    return "";
  }
}
</script>
