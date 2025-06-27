<template>
  <v-card title="Inscription reçue">
    <v-card-text>
      Nous avons bien reçu votre demande d'inscription, et nous vous en
      remercions !

      <v-alert type="warning" v-if="aStatuer.length" class="my-2">
        <v-row>
          <v-col cols="8">
            Nous n'avons pas encore statué pour ces demandes et nous reviendrons
            vers vous au plus vite.
          </v-col>
          <v-col align-self="center" class="text-right">
            <v-chip v-for="participant in aStatuer">
              {{ participant.Personne.Prenom }} :
              {{ Camps.label(participant.Camp) }}
            </v-chip>
          </v-col>
        </v-row>
      </v-alert>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { StatutParticipant, type DossierExt } from "../logic/api";
import { Camps } from "@/utils";

const props = defineProps<{
  dossier: DossierExt;
}>();

const aStatuer = computed(() =>
  (props.dossier.Participants || []).filter(
    (p) => p.Participant.Statut == StatutParticipant.AStatuer
  )
);
</script>

<style scoped></style>
