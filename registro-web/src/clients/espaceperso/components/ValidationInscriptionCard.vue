<template>
  <v-card title="Inscription confirmée">
    <v-card-text>
      <div class="mb-4">
        Nous avons bien reçu votre inscription, et nous vous en remercions.
        Voici son statut :
      </div>

      <!-- inscrits -->
      <v-alert
        v-if="inscrits.length"
        color="lime-lighten-1"
        class="my-2"
        elevation="2"
        :icon="Formatters.statutParticipant(StatutParticipant.Inscrit).icon"
      >
        <v-row>
          <v-col align-self="center" class="text-center">
            <v-chip v-for="participant in inscrits" class="my-1">
              {{ participant.Personne.Prenom }} :
              {{ Camps.label(participant.Camp) }}
            </v-chip>
          </v-col>
          <v-col cols="7"
            >Nous sommes ravis de vous confirmer
            {{
              inscrits.length == 1
                ? "cette participation"
                : "ces participations"
            }}
            et nous nous réjouissons des moments à passer ensemble !</v-col
          >
        </v-row>
      </v-alert>

      <!-- attente -->
      <v-alert
        v-if="attente.length"
        color="orange-lighten-3"
        class="my-2"
        elevation="2"
        :icon="
          Formatters.statutParticipant(StatutParticipant.AttenteCampComplet)
            .icon
        "
      >
        <v-row>
          <v-col align-self="center" class="text-center">
            <v-chip v-for="participant in attente" class="my-1">
              {{ participant.Personne.Prenom }} :
              {{ Camps.label(participant.Camp) }}
            </v-chip>
          </v-col>
          <v-col cols="7"
            >Malheureusement, nous avons dû placer
            {{ attente.length == 1 ? "cette demande" : "ces demandes" }}
            en liste d'attente. Nous reviendrons vers vous si une place se
            libère.</v-col
          >
        </v-row>
      </v-alert>

      <!-- a statuer -->
      <v-alert
        v-if="aStatuer.length"
        color="grey-lighten-1"
        class="my-2"
        elevation="2"
        :icon="Formatters.statutParticipant(StatutParticipant.AStatuer).icon"
      >
        <v-row>
          <v-col align-self="center" class="text-center">
            <v-chip v-for="participant in aStatuer" class="my-1">
              {{ participant.Personne.Prenom }} :
              {{ Camps.label(participant.Camp) }}
            </v-chip>
          </v-col>
          <v-col cols="7">
            Nous n'avons pas encore statué pour
            {{ aStatuer.length == 1 ? "cette demande" : "ces demandes" }}
            et nous reviendrons vers vous au plus vite.
          </v-col>
        </v-row>
      </v-alert>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { StatutParticipant, type DossierExt } from "../logic/api";
import { Camps, Formatters } from "@/utils";

const props = defineProps<{
  dossier: DossierExt;
}>();

const inscrits = computed(() =>
  (props.dossier.Participants || []).filter(
    (p) => p.Participant.Statut == StatutParticipant.Inscrit
  )
);
const attente = computed(() =>
  (props.dossier.Participants || []).filter(
    (p) =>
      p.Participant.Statut != StatutParticipant.Inscrit &&
      p.Participant.Statut != StatutParticipant.AStatuer
  )
);
const aStatuer = computed(() =>
  (props.dossier.Participants || []).filter(
    (p) => p.Participant.Statut == StatutParticipant.AStatuer
  )
);
</script>

<style scoped></style>
