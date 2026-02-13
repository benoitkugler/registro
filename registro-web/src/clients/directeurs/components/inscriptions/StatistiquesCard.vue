<template>
  <v-card subtitle="Statistiques" width="300px">
    <v-card-text>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Inscrits </v-col>
        <v-col cols="auto">
          <v-badge color="primary" inline :content="inscrits.length"> </v-badge>
        </v-col>
      </v-row>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Garçons/Filles </v-col>
        <v-col cols="auto">
          <v-badge
            color="primary"
            inline
            :content="`${
              inscrits.filter((p) => p.Personne.Sexe == Sexe.Man).length
            } /  ${
              inscrits.filter((p) => p.Personne.Sexe == Sexe.Woman).length
            }`"
          >
          </v-badge>
        </v-col>
      </v-row>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Anniversaires </v-col>
        <v-col cols="auto">
          <v-badge
            color="amber-lighten-1"
            inline
            :content="inscrits.filter((p) => p.HasBirthday).length"
          >
          </v-badge>
        </v-col>
      </v-row>
      <v-divider thickness="1"></v-divider>
      <v-row class="my-1" no-gutters justify="space-between">
        <v-col> Liste d'attente </v-col>
        <v-col cols="auto">
          <v-badge
            color="orange"
            inline
            :content="props.participants.length - inscrits.length"
          >
          </v-badge>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { Sexe, StatutParticipant, type ParticipantExt } from "../../logic/api";

const props = defineProps<{
  participants: ParticipantExt[];
}>();

const inscrits = computed(() =>
  props.participants.filter(
    (p) => p.Participant.Statut == StatutParticipant.Inscrit
  )
);
</script>
