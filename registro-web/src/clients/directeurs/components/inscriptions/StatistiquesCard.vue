<template>
  <v-card subtitle="Statistiques" width="300px">
    <v-card-text>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Inscrits </v-col>
        <v-col cols="auto">
          <v-badge color="primary" inline :content="props.statistiques.Valides">
          </v-badge>
        </v-col>
      </v-row>
      <v-row
        class="my-1"
        no-gutters
        justify="space-between"
        v-if="props.statistiques.ValidesSuisses"
      >
        <v-col> Dont suisses </v-col>
        <v-col cols="auto">
          <v-badge
            color="primary"
            inline
            :content="props.statistiques.ValidesSuisses"
          >
          </v-badge>
        </v-col>
      </v-row>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Garçons/Filles </v-col>
        <v-col cols="auto">
          <v-badge
            color="primary"
            inline
            :content="`${
              props.statistiques.Valides - props.statistiques.ValidesFilles
            } /  ${props.statistiques.ValidesFilles}`"
          >
          </v-badge>
        </v-col>
      </v-row>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Anniversaires </v-col>
        <v-col cols="auto">
          <v-badge color="amber-lighten-1" inline :content="birthdays">
          </v-badge>
        </v-col>
      </v-row>
      <v-divider thickness="1"></v-divider>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> A statuer </v-col>
        <v-col cols="auto">
          <v-badge color="grey" inline :content="props.statistiques.AStatuer">
          </v-badge>
        </v-col>
      </v-row>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Liste d'attente </v-col>
        <v-col cols="auto">
          <v-badge
            color="orange"
            inline
            :content="props.statistiques.ListeAttente"
          >
          </v-badge>
        </v-col>
      </v-row>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Refus définitif </v-col>
        <v-col cols="auto">
          <v-badge color="red" inline :content="props.statistiques.Refus">
          </v-badge>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import {
  Sexe,
  StatutParticipant,
  type ParticipantExt,
  type StatistiquesInscrits,
} from "../../logic/api";

const props = defineProps<{
  participants: ParticipantExt[]; // for birthday
  statistiques: StatistiquesInscrits;
}>();

const birthdays = computed(
  () =>
    props.participants.filter(
      (p) => p.Participant.Statut == StatutParticipant.Inscrit && p.HasBirthday
    ).length
);
</script>
